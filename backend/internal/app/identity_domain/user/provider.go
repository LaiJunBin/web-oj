package user

import (
	"judge-backend/internal/app/shared/types/provider"
	"judge-backend/internal/app/identity_domain/user/adapter/controller"
	"judge-backend/internal/app/identity_domain/user/adapter/presenter"
	"judge-backend/internal/app/identity_domain/user/application"
	createuser "judge-backend/internal/app/identity_domain/user/application/usecase/create_user"
	deleteuser "judge-backend/internal/app/identity_domain/user/application/usecase/delete_user"
	getuser "judge-backend/internal/app/identity_domain/user/application/usecase/get_user"
	importusers "judge-backend/internal/app/identity_domain/user/application/usecase/import_users"
	listrole "judge-backend/internal/app/identity_domain/user/application/usecase/list_role"
	listuser "judge-backend/internal/app/identity_domain/user/application/usecase/list_user"
	updateuser "judge-backend/internal/app/identity_domain/user/application/usecase/update_user"
	"judge-backend/internal/app/identity_domain/user/infrastructure/repository"
	"judge-backend/internal/app/identity_domain/user/infrastructure/service/hashservice"
	"judge-backend/internal/app/identity_domain/user/infrastructure/service/query/rolequery"
	"judge-backend/internal/app/identity_domain/user/infrastructure/service/query/userquery"
	"judge-backend/internal/app/identity_domain/user/infrastructure/service/user/roleconverter"

	adduserstogroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/add_users_to_group"
	creategroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/create_group"
	deletegroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/delete_group"
	getgroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/get_group"
	listgroups "judge-backend/internal/app/identity_domain/user/application/usecase/group/list_groups"
	removeusersfromgroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/remove_users_from_group"
	updategroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/update_group"

)

func NewProvider() *provider.Provider {
	return provider.New(provider.Config{
		Infrastructures: []interface{}{
			repository.NewUserRepository,
			repository.NewGroupRepository,
			userquery.NewUserQueryService,
			rolequery.NewRoleQueryService,
			hashservice.NewHashService,
			roleconverter.NewRoleConverter,
		},
		Controllers: []interface{}{
			controller.NewUserController,
			controller.NewGroupController,
		},
		Presenters: []interface{}{
			presenter.NewUserPresenter,
		},
		Usecases: []interface{}{
			listrole.NewUsecase,
			listuser.NewUsecase,
			getuser.NewUsecase,
			createuser.NewUsecase,
			updateuser.NewUsecase,
			deleteuser.NewUsecase,
			importusers.NewUsecase,

			creategroup.New,
			getgroup.New,
			listgroups.New,
			updategroup.New,
			deletegroup.New,
			adduserstogroup.New,
			removeusersfromgroup.New,

		},
		Applications: []interface{}{
			application.NewUserApplication,
			application.NewGroupApplication,
		},
	})
}
