package repository

type ParticipantRepository interface {
	AddUsers(contestID int32, userIDs []int) error
	RemoveUsers(contestID int32, userIDs []int) error
	AddGroups(contestID int32, groupIDs []int) error
	RemoveGroups(contestID int32, groupIDs []int) error
	GetUserIDs(contestID int32) ([]int, error)
	GetGroupIDs(contestID int32) ([]int, error)
}
