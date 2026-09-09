package service

import (
	"bytes"
)

type ProjectSubmitPublisher interface {
	// Publish validates the dynamic fields synchronously and, only if they are
	// valid, emits the submission event. An error means nothing was published —
	// the caller must surface it to the contestant rather than answer success,
	// so a bad submission is rejected up front instead of vanishing later.
	Publish(
		userID int32,
		contestProjectID int32,
		isPartial bool,
		gradingItemIDs []int32,
		project *bytes.Buffer,
		dynamicFields map[string]string,
	) error
}
