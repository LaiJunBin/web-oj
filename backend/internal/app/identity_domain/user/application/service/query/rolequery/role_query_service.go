package rolequery

type Role struct {
	ID   int32
	Name string
}

type RoleQueryService interface {
	ListRole() ([]Role, error)
}
