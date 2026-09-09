package service

import (
	"time"

	exportbundle "judge-backend/internal/app/portability_domain/bundle/application/usecase/export_bundle"
)

type systemClock struct{}

// NewClock supplies the manifest's provenance timestamp. Injected so tests can
// pin the one non-deterministic input to a composite export.
func NewClock() exportbundle.Clock { return systemClock{} }

func (systemClock) NowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }
