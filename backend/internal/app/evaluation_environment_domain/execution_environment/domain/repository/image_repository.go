package repository

import (
	"errors"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
)

var (
	ErrCreatedAtMustNotBeZero = errors.New("when updating an existing image, created_at must not be zero")
	ErrCreatedAtMustBeZero    = errors.New("when creating a new image, created_at must be zero")
	// ErrImageTagAlreadyExists is CreateExclusive's duplicate verdict: the
	// tag's unique index rejected the insert because another (possibly
	// concurrent) request already owns it.
	ErrImageTagAlreadyExists = errors.New("an image with this tag already exists")
)

type ImageRepository interface {
	GetAll() ([]*image.Image, error)
	FindByTag(tag string) (*image.Image, error)
	Save(image *image.Image) error
	// CreateExclusive INSERTS a brand-new image row, never updates: the
	// database's unique tag index is the arbiter, so of N concurrent
	// creates exactly one wins and every loser gets
	// ErrImageTagAlreadyExists. Save's read-then-upsert cannot express this
	// — a loser whose read ran after the winner's commit would silently
	// UPDATE the winner's row instead of being told it lost. The returned
	// ID is the caller's stable ownership token: compensation must delete
	// through DeleteIfOwned(token), never by tag, because the tag can be
	// deleted and re-created by another request while the caller is still
	// in flight — by then the tag resolves to the NEW owner's row.
	CreateExclusive(image *image.Image) (int, error)
	// DeleteIfOwned deletes exactly the row the token identifies. A stale
	// token (the row was already deleted, e.g. by the user, before this
	// compensation ran) is a no-op — it can never touch a row a later
	// CreateExclusive now owns, since that row has a different ID.
	//
	// The returned bool reports whether the row was actually deleted BY THIS
	// CALL: true means the token's row still existed a moment ago, i.e. the
	// caller still owned the tag; false means someone else already removed
	// it. Callers use this as the ownership verdict for the shared,
	// tag-derived project path — only a true verdict may touch the files
	// (same MatchedCount-as-ownership pattern as the judge domain's
	// conditional judgment updates).
	DeleteIfOwned(id int) (bool, error)
	Delete(image *image.Image) error
}
