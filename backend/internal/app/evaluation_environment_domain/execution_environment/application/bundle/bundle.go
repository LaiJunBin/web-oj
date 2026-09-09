// Package bundle is the on-disk form of an evaluation image.
//
// A project bundle names the image it needs but never carries it, because an
// image is built on the machine that runs the contest. This package is what
// carries it, and it does so in one of two shapes the operator picks between:
//
//   - ModeSource holds the build context — a Dockerfile and whatever it
//     copies. Kilobytes, readable, reviewable in a diff. The receiving machine
//     builds it, which takes minutes and needs the network to pull base
//     images.
//   - ModeArchive holds the output of docker save. Hundreds of megabytes, but
//     the receiving machine only loads it: no build, no network, ready at
//     once. This is the shape an air-gapped venue needs.
//
// Neither is the default. The choice depends on the venue's network, which the
// exporting installation cannot know.
package bundle

import bundlecodec "judge-backend/internal/app/shared/types/bundle_codec"

// FormatVersion is the schema version written into every image bundle.
// Any schema change bumps it, additive ones included — see bundlecodec.
const FormatVersion = 1

// KindImage marks a bundle carrying one image.
const KindImage = "image"

// The two shapes an image bundle can take.
const (
	ModeSource  = "source"
	ModeArchive = "archive"
)

// File names inside an image bundle's zip.
const (
	// ImageFile is the metadata and the fingerprint.
	ImageFile = "image.json"
	// ContextFile is the build context, present when Mode is ModeSource.
	ContextFile = "context.zip"
	// ArchiveFile is the docker save output, present when Mode is ModeArchive.
	ArchiveFile = "image.tar"
)

// Image is the definition of one evaluation image.
type Image struct {
	FormatVersion int    `json:"formatVersion"`
	Kind          string `json:"kind"`

	// Tag is both the identity and the destination: an image imported here
	// takes the same tag it had, because that is what the projects referencing
	// it name.
	Tag string `json:"tag"`

	MountPath   string `json:"mountPath"`
	Description string `json:"description"`

	Mode string `json:"mode"`

	// Fingerprint identifies the CONTENT, so the importer can tell "I already
	// have this" from "I have something else under this name".
	//
	// Without it the only comparison available is the tag, and a tag says
	// nothing about what is inside — importing would either refuse every
	// image the venue already has, or silently overwrite one that several
	// live contests depend on. Its meaning depends on Mode: for a source
	// bundle it covers the build context's files, for an archive it is the
	// image digest the daemon reports.
	Fingerprint string `json:"fingerprint"`
}

// IsReady is deliberately absent: whether an image is built is a fact about a
// particular machine, not a property of the image, and importing one must not
// claim readiness the receiving daemon has not earned.

func Marshal(i *Image) ([]byte, error) {
	return bundlecodec.Marshal(i)
}

func Unmarshal(data []byte) (*Image, error) {
	var i Image
	if err := bundlecodec.Unmarshal(data, &i); err != nil {
		return nil, err
	}
	return &i, nil
}
