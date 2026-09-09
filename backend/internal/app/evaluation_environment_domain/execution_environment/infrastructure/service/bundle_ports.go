package service

import (
	"context"
	"io"

	exportimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/export_image"
	importimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/import_image"
	"judge-backend/internal/app/shared/model"

	"github.com/docker/docker/client"
	"gorm.io/gorm"
)

// The daemon and database adapters behind the image bundle's ports. Kept
// behind interfaces so the usecases can be exercised without a Docker daemon —
// the alternative is that the only test of an archive import is a feature test
// that pulls real layers.

type dockerImages struct{ client client.APIClient }

// NewImageArchiver and NewImageLoader are the same adapter under two names:
// export needs save+digest, import needs load+digest, and neither should be
// handed the half it has no business calling.
func NewImageArchiver(c client.APIClient) exportimage.ImageArchiver { return &dockerImages{client: c} }

func NewImageLoader(c client.APIClient) importimage.ImageLoader { return &dockerImages{client: c} }

func (d *dockerImages) Save(ctx context.Context, tag string) (io.ReadCloser, error) {
	return d.client.ImageSave(ctx, []string{tag})
}

func (d *dockerImages) Load(ctx context.Context, tar io.Reader) error {
	resp, err := d.client.ImageLoad(ctx, tar, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Drained rather than ignored: the daemon streams its progress here, and
	// leaving it unread can stall the request on some transports.
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

func (d *dockerImages) Digest(ctx context.Context, tag string) (string, error) {
	inspect, _, err := d.client.ImageInspectWithRaw(ctx, tag)
	if err != nil {
		return "", err
	}
	// The image ID is the content address of the built image, which is what a
	// fingerprint needs. RepoDigests would be emptier and registry-dependent.
	return inspect.ID, nil
}

type tagUsage struct{ db *gorm.DB }

// NewTagUsage reports which projects reference an image tag, so a conflict can
// tell the operator what forcing would change.
func NewTagUsage(db *gorm.DB) importimage.TagUsage { return &tagUsage{db: db} }

func (t *tagUsage) ProjectsUsing(tag string) ([]string, error) {
	var names []string
	if err := t.db.Model(&model.Project{}).
		Where("image_tag = ? OR test_image_tag = ?", tag, tag).
		Order("id ASC").
		Pluck("name", &names).Error; err != nil {
		return nil, err
	}
	return names, nil
}
