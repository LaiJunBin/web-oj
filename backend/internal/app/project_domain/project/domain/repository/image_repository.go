package repository

type ImageRepository interface {
	Exists(tag string) (bool, error)
}
