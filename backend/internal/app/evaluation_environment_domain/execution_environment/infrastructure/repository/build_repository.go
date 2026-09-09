package repository

import (
	"sync"

	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/buildloglevel"
	buildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/log"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type BuildRepository struct {
	db *gorm.DB
	// mu guards buildMap: the queue workers executing builds, the async
	// cancel handler, and API goroutines (EnqueueBuildJob / the cancel
	// usecase) all read and write it concurrently.
	mu       sync.Mutex
	buildMap map[string]*build.Build
}

func NewBuildRepository(
	db *gorm.DB,
) repository.BuildRepository {
	return &BuildRepository{
		db:       db,
		buildMap: map[string]*build.Build{},
	}
}

// tagToImageID resolves a tag to its LIVE image row first, and only falls
// back to soft-deleted rows when no live row exists.
//
// The order is the whole point. A force import USED TO replace an image by
// soft-deleting the old row and inserting a new one, so the same tag then had
// both (it now replaces in place, and deletion renames the dead row's tag).
// The old code resolved with an unscoped unordered First — which returns
// the LOWEST id, the soft-deleted row — so every build after a force import
// was silently attached to the dead image: the builds list (which resolves the
// live row) showed nothing, and rebuild / update-with-context walked
// FindLastBuild into the dead image's builds and 500ed. The fallback is now
// defensive only: soft-deleted rows carry a renamed tag (model.DeletedImageTag
// frees the tag on MySQL's non-partial unique index), so an original-tag
// lookup no longer resolves dead rows, and delete_image cancels the build
// BEFORE deleting the row (the scoped lookup above hits directly). It stays
// as a safety net.
func (r *BuildRepository) tagToImageID(tag string) (int, error) {
	image := &model.Image{}
	err := r.db.First(image, "tag = ?", tag).Error
	if err == nil {
		return image.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return 0, err
	}

	// Newest first: a tag can be deleted and recreated many times, and the
	// builds being resolved belong to the most recent incarnation.
	if err := r.db.Unscoped().Order("id DESC").First(image, "tag = ?", tag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, repository.ErrorResolveBuildTag
		}
		return 0, err
	}

	return image.ID, nil
}

// imageIDToTag reads unscoped on purpose, unlike the tag lookup above: an id
// is unambiguous, and this only labels a build with its image's tag. Scoped,
// it turned a build row belonging to a soft-deleted image into a hard error —
// which is what escalated the force-import misattachment from "invisible
// build" to "rebuild returns 500". A soft-deleted image's tag is renamed
// (model.DeletedImageTag), so build history for a dead image shows
// "{tag}__deleted_{id}" — deliberately: it marks the build as belonging to a
// deleted incarnation, and the suffix keeps imagetag validation passing.
func (r *BuildRepository) imageIDToTag(imageID int) (string, error) {
	image := &model.Image{}
	if err := r.db.Unscoped().First(image, "id = ?", imageID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", repository.ErrorResolveImageTag
		}
		return "", err
	}

	return image.Tag, nil
}

func (r *BuildRepository) toModelBuild(build *build.Build) (*model.Build, error) {
	id, err := r.tagToImageID(build.Tag())
	if err != nil {
		return nil, err
	}

	b := &model.Build{}
	if err := r.db.First(b, "image_id = ? AND batch = ?", id, build.Batch().Value()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrorResolveBuildModel
		}
		return nil, err
	}

	return b, nil
}

func (r *BuildRepository) toDomainBuild(b *model.Build) (*build.Build, error) {
	t, err := r.imageIDToTag(b.ImageID)
	if err != nil {
		return nil, err
	}

	tag, err := imagetag.New(t)
	if err != nil {
		return nil, repository.ErrorResolveBuildTag
	}

	status, err := buildstatus.New(b.Status)
	if err != nil {
		return nil, err
	}

	config := &build.BuildConfig{
		Tag:    tag,
		Batch:  batch.MustNew(b.Batch),
		Status: status,
	}

	build := build.New(config)
	for _, l := range b.Logs {
		level, err := buildloglevel.New(l.Level)
		if err != nil {
			return nil, err
		}

		log := buildlog.New(level, l.Message, l.CreatedAt)
		build.AddLog(log)
	}

	return build, nil
}

func (r *BuildRepository) GetMaxBatch(tag string) (int, error) {
	id, err := r.tagToImageID(tag)
	if err != nil {
		return 0, err
	}

	var maxBatch int
	if err := r.db.Model(&model.Build{}).Where("image_id = ?", id).Select("MAX(batch)").Scan(&maxBatch).Error; err != nil {
		return 0, nil
	}

	return maxBatch, nil
}

func (r *BuildRepository) saveKey(tag string) string {
	return tag
}

func (r *BuildRepository) FindByTagAndBatch(tag string, batch int) (*build.Build, error) {
	r.mu.Lock()
	if b, ok := r.buildMap[r.saveKey(tag)]; ok {
		if b.Batch().Value() == batch {
			r.mu.Unlock()
			return b, nil
		}
	}
	r.mu.Unlock()

	id, err := r.tagToImageID(tag)
	if err != nil {
		return nil, nil
	}

	build := &model.Build{}
	if err := r.db.Preload("Logs").First(build, "image_id = ? AND batch = ?", id, batch).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return r.toDomainBuild(build)
}

func (r *BuildRepository) FindLastBuild(tag string) (*build.Build, error) {
	key := r.saveKey(tag)
	r.mu.Lock()
	if b, ok := r.buildMap[key]; ok {
		r.mu.Unlock()
		return b, nil
	}
	r.mu.Unlock()

	batch, err := r.GetMaxBatch(tag)
	if err != nil {
		return nil, nil
	}

	return r.FindByTagAndBatch(tag, batch)
}

// SavePendingExclusive atomically checks-and-registers a pending build; see
// the interface doc comment for the TOCTOU it closes.
func (r *BuildRepository) SavePendingExclusive(build *build.Build) error {
	key := r.saveKey(build.Tag())
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.buildMap[key]; ok && (existing.IsPending() || existing.IsBuilding()) {
		return repository.ErrorBuildAlreadyInProgress
	}
	r.buildMap[key] = build
	return nil
}

func (r *BuildRepository) Commit(build *build.Build) error {
	b, err := r.toModelBuild(build)
	if err != nil {
		imageID, err := r.tagToImageID(build.Tag())
		if err != nil {
			return err
		}
		b = &model.Build{
			ImageID: imageID,
			Batch:   build.Batch().Value(),
		}
	}

	b.Status = build.Status().Value()
	b.Logs = []model.BuildLog{}

	for _, l := range build.Logs() {
		b.Logs = append(b.Logs, model.BuildLog{
			Level:     l.Level().Value(),
			Message:   l.Message(),
			CreatedAt: l.Time(),
		})
	}

	tx := r.db.Begin()
	if err := tx.Where("build_id = ?", b.ID).Delete(&model.BuildLog{}).Error; err != nil {
		return err
	}

	if err := tx.Save(b).Error; err != nil {
		return repository.ErrorSaveBuild
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	r.unregisterPending(build)

	return nil
}

// ReleasePending is the public identity-guarded release for reservations
// whose setup/enqueue failed; see the interface doc comment.
func (r *BuildRepository) ReleasePending(build *build.Build) {
	r.unregisterPending(build)
}

// unregisterPending removes build's map entry, but only if the entry still
// IS this build. By the time an old build's deferred commit runs, a NEWER
// pending build of the same tag can already occupy the key — an
// unconditional delete would orphan it (its event would find nothing, and
// it could never be cancelled).
func (r *BuildRepository) unregisterPending(build *build.Build) {
	r.mu.Lock()
	if r.buildMap[r.saveKey(build.Tag())] == build {
		delete(r.buildMap, r.saveKey(build.Tag()))
	}
	r.mu.Unlock()
}

func (r *BuildRepository) FindCurrentBuildByTag(tag string) *build.Build {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.buildMap[r.saveKey(tag)]; ok {
		return b
	}

	return nil
}
