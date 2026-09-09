package repository

import (
	"errors"
	"strings"
	"time"

	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/provider"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type ImageRepository struct {
	db          *gorm.DB
	storagePath string
}

func NewImageRepository(
	db *gorm.DB,
	storagePathProvider provider.StoragePathProvider,
) repository.ImageRepository {
	return &ImageRepository{
		db:          db,
		storagePath: storagePathProvider.GetStoragePath(),
	}
}

func (r *ImageRepository) toDomainImage(img *model.Image) (*image.Image, error) {
	domainImage, err := image.New(&image.Config{
		Tag:         img.Tag,
		MountPath:   img.MountPath,
		Description: img.Description,
		StoragePath: r.storagePath,
	})
	if err != nil {
		return nil, err
	}

	if img.IsReady {
		domainImage.MarkAsReady()
	}

	domainImage.SetCreatedAt(img.CreatedAt)
	domainImage.SetUpdatedAt(img.UpdatedAt)

	return domainImage, nil
}

func (r *ImageRepository) getID(image *image.Image) (int, error) {
	img := &model.Image{}
	if err := r.db.First(img, "tag = ?", image.Tag().String()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}

		return 0, err
	}

	return img.ID, nil
}

func (r *ImageRepository) GetAll() ([]*image.Image, error) {
	images := make([]*model.Image, 0)
	if err := r.db.Find(&images).Error; err != nil {
		return nil, err
	}

	domainImages := make([]*image.Image, 0, len(images))
	for _, img := range images {
		domainImage, err := r.toDomainImage(img)
		if err != nil {
			return nil, err
		}
		domainImages = append(domainImages, domainImage)
	}

	return domainImages, nil
}

func (r *ImageRepository) FindByTag(tag string) (*image.Image, error) {
	img := &model.Image{}
	if err := r.db.First(img, "tag = ?", tag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomainImage(img)
}

func (r *ImageRepository) Save(image *image.Image) error {
	id, err := r.getID(image)
	if err != nil {
		return err
	}

	if id != 0 && image.CreatedAt().IsZero() {
		return repository.ErrCreatedAtMustNotBeZero
	}

	if id == 0 && !image.CreatedAt().IsZero() {
		return repository.ErrCreatedAtMustBeZero
	}

	img := &model.Image{
		ID:          id,
		Tag:         image.Tag().String(),
		MountPath:   image.MountPath().String(),
		Description: image.Description(),
		IsReady:     image.IsReady(),
	}
	img.CreatedAt = image.CreatedAt()

	if err := r.db.Save(img).Error; err != nil {
		return err
	}

	image.SetCreatedAt(img.CreatedAt)
	image.SetUpdatedAt(img.UpdatedAt)

	return nil
}

// isDuplicateKeyError centralizes the dialect-specific duplicate detection
// so no caller ever guesses at raw database errors: gorm's translated
// sentinel, MySQL error 1062, or sqlite's UNIQUE-constraint message.
func isDuplicateKeyError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var mysqlErr *mysqldriver.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// CreateExclusive INSERTS only — see the interface doc comment. The tag's
// unique index arbitrates concurrent creates; a duplicate maps to
// ErrImageTagAlreadyExists. The new row's ID is returned as the caller's
// ownership token for DeleteIfOwned.
func (r *ImageRepository) CreateExclusive(image *image.Image) (int, error) {
	if !image.CreatedAt().IsZero() {
		return 0, repository.ErrCreatedAtMustBeZero
	}

	img := &model.Image{
		Tag:         image.Tag().String(),
		MountPath:   image.MountPath().String(),
		Description: image.Description(),
		IsReady:     image.IsReady(),
	}

	if err := r.db.Create(img).Error; err != nil {
		if isDuplicateKeyError(err) {
			return 0, repository.ErrImageTagAlreadyExists
		}
		return 0, err
	}

	image.SetCreatedAt(img.CreatedAt)
	image.SetUpdatedAt(img.UpdatedAt)

	return img.ID, nil
}

// softDeleteFreeingTag flips deleted_at AND renames the tag in ONE UPDATE:
// on MySQL the unique tag index has no partial-index WHERE clause, so a
// soft-deleted row that keeps its tag blocks recreating that tag forever
// (see model.Image.Tag). Unscoped plus the explicit deleted_at IS NULL guard
// make this conditional update the single arbiter — RowsAffected stays the
// ownership verdict DeleteIfOwned is built on. The new tag is composed in Go
// (model.DeletedImageTag) so no dialect-specific CONCAT is needed; setting
// deleted_at by hand is exactly what gorm's own soft delete does.
func (r *ImageRepository) softDeleteFreeingTag(id int, tag string) (int64, error) {
	res := r.db.Model(&model.Image{}).Unscoped().
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"tag":        model.DeletedImageTag(tag, id),
		})
	return res.RowsAffected, res.Error
}

// DeleteIfOwned deletes by primary key only — see the interface doc
// comment. Unlike Delete, it NEVER re-resolves the tag: the row the token
// identifies either still belongs to the caller (deleted) or is already
// gone (no-op); a row a later create owns has a different ID and is
// unreachable from here. RowsAffected is the ownership verdict: > 0 means
// this call deleted the row, so the caller held the tag until this moment.
func (r *ImageRepository) DeleteIfOwned(id int) (bool, error) {
	if id == 0 {
		return false, nil
	}
	// Scoped read: a row someone else already soft-deleted is not ours, and
	// the verdict must say so. The tag read here cannot go stale before the
	// conditional UPDATE below — a live row's tag is never updated in this
	// codebase (update_image leaves it alone, force import replaces
	// in-place precisely to keep it).
	img := &model.Image{}
	if err := r.db.First(img, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	n, err := r.softDeleteFreeingTag(id, img.Tag)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *ImageRepository) Delete(image *image.Image) error {
	id, err := r.getID(image)
	if err != nil {
		return err
	}
	if id == 0 {
		// No live row for this tag — the delete is an idempotent no-op.
		// (Previously this fell through to gorm's ErrMissingWhereClause
		// guard; unreachable from the usecase, which resolves the row
		// first.)
		return nil
	}

	_, err = r.softDeleteFreeingTag(id, image.Tag().String())
	return err
}
