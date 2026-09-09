package jsonpresenter

import (
	"errors"

	"judge-backend/api"
	previewbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/preview_bundle"
)

type Presenter[I any, O any] interface {
	Output(data *I) O
	Error(err error) O
}

type PreviewBundlePresenter Presenter[previewbundle.Output, api.PreviewBundleResponseObject]

type previewBundlePresenter struct{}

func NewPreviewBundlePresenter() PreviewBundlePresenter { return &previewBundlePresenter{} }

func (p *previewBundlePresenter) Output(output *previewbundle.Output) api.PreviewBundleResponseObject {
	return api.PreviewBundle200JSONResponse{
		Success: true,
		Data:    toAPIPreview(output),
	}
}

func (p *previewBundlePresenter) Error(err error) api.PreviewBundleResponseObject {
	var invalid *previewbundle.InvalidBundleError
	if errors.As(err, &invalid) {
		return api.PreviewBundle400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			Details: invalid.Details,
		}
	}

	return api.PreviewBundle500JSONResponse{
		Success: false,
		Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
	}
}

func toAPIPreview(output *previewbundle.Output) api.AdminBundleSpecBundlePreview {
	preview := api.AdminBundleSpecBundlePreview{
		FormatVersion: int32(output.FormatVersion),
		Images:        make([]api.AdminBundleSpecBundleImageItem, 0, len(output.Images)),
		Projects:      make([]api.AdminBundleSpecBundleProjectItem, 0, len(output.Projects)),
		Blockers:      output.Blockers,
	}

	for _, i := range output.Images {
		preview.Images = append(preview.Images, api.AdminBundleSpecBundleImageItem{
			Tag:            i.Tag,
			Mode:           i.Mode,
			Status:         api.AdminBundleSpecBundleItemStatus(i.Status),
			UsedByProjects: i.UsedByProjects,
		})
	}

	for _, p := range output.Projects {
		preview.Projects = append(preview.Projects, api.AdminBundleSpecBundleProjectItem{
			Name:             p.Name,
			Status:           api.AdminBundleSpecBundleItemStatus(p.Status),
			MissingImageTags: p.MissingImageTags,
		})
	}

	if output.Contest != nil {
		preview.Contest = &api.AdminBundleSpecBundleContestItem{
			Name:               output.Contest.Name,
			ProjectSlots:       output.Contest.ProjectSlots,
			UnresolvedProjects: output.Contest.UnresolvedProjects,
		}
	}

	return preview
}
