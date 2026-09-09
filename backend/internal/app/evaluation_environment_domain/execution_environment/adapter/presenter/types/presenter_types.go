package types

import (
	"context"
	"judge-backend/api"
	cancelbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/cancel_build"
	createimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/create_image"
	deleteimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/delete_image"
	exportimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/export_image"
	getbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/get_build"
	getimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/get_image"
	importimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/import_image"
	listbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/list_build"
	listimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/list_image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/rebuild"
	streambuildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/stream_build_log"
	updateimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/update_image"
)

type Presenter[I any, O any] interface {
	Output(data *I) O
	Error(err error) O
}

type ListImagePresenterType Presenter[listimage.Output, api.ListImageResponseObject]
type GetImagePresenterType Presenter[getimage.Output, api.GetImageResponseObject]
type CreateImagePresenterType Presenter[createimage.Output, api.CreateImageResponseObject]
type UpdateImagePresenterType Presenter[updateimage.Output, api.UpdateImageResponseObject]
type DeleteImagePresenterType Presenter[deleteimage.Output, api.DeleteImageResponseObject]
type ListBuildPresenterType Presenter[listbuild.Output, api.ListBuildResponseObject]
type GetBuildPresenterType Presenter[getbuild.Output, api.GetBuildResponseObject]
type CancelBuildPresenterType Presenter[cancelbuild.Output, api.CancelBuildResponseObject]
type RebuildPresenterType Presenter[rebuild.Output, api.RebuildResponseObject]
type StreamBuildLogPresenterType interface {
	Output(ctx context.Context, data *streambuildlog.Output) api.StreamLogsResponseObject
	Error(err error) api.StreamLogsResponseObject
}

type ExportImagePresenterType Presenter[exportimage.Output, api.ExportImageResponseObject]
type ImportImagePresenterType Presenter[importimage.Output, api.ImportImageResponseObject]
