package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"judge-backend/internal/app/judge_domain/judge/domain/provider"
	"judge-backend/internal/app/judge_domain/judge/domain/service"
	sharedservice "judge-backend/internal/app/shared/service"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/submissiondynamicfield"
)

type SubmissionService struct {
	storagePath string
	fss         sharedservice.FileSystemService
	zip         sharedservice.ZipService
	logger      sharedtypes.Logger
}

func NewSubmissionService(
	storagePathProvider provider.StoragePathProvider,
	fss sharedservice.FileSystemService,
	zip sharedservice.ZipService,
	logger sharedtypes.Logger,
) service.SubmissionService {
	return &SubmissionService{
		storagePath: storagePathProvider.GetStoragePath(),
		fss:         fss,
		zip:         zip,
		logger:      logger,
	}
}

func (s *SubmissionService) dirPath(submissionID string) string {
	return fmt.Sprintf("%s/submissions/%s", s.storagePath, submissionID)
}

func (s *SubmissionService) projectPath(submissionID string) string {
	return fmt.Sprintf("%s/project", s.dirPath(submissionID))
}

func (s *SubmissionService) filesPath(submissionID string) string {
	return fmt.Sprintf("%s/files", s.dirPath(submissionID))
}

func (s *SubmissionService) metadataPath(submissionID string) string {
	return fmt.Sprintf("%s/metadata.json", s.dirPath(submissionID))
}

// archivePath is the zip the contestant actually uploaded, kept beside the
// extracted tree rather than inside it. The judge runtime copies only
// <submission>/project into the graded container, so an archive at this
// level never reaches the contestant's container or the web root.
//
// Mirrors the grading-project layout, which already stores project.zip
// beside its files/ directory.
func (s *SubmissionService) archivePath(submissionID string) string {
	return fmt.Sprintf("%s/project.zip", s.dirPath(submissionID))
}

// errEmptySubmissionID guards every path-building method: an empty ID would
// silently target storage/submissions/ itself (catastrophic for cleanup).
var errEmptySubmissionID = errors.New("submission ID must not be empty")

// containedPath joins base and name, then rejects the result unless it stays
// inside base. name is a project-defined dynamic field key that becomes a
// filename, so without this a "../" segment could escape the submission's files
// directory and overwrite files reachable by the server process.
func containedPath(base string, name string) (string, error) {
	cleanBase := filepath.Clean(base)
	joined := filepath.Join(cleanBase, name)
	if joined != cleanBase && !strings.HasPrefix(joined, cleanBase+string(filepath.Separator)) {
		return "", fmt.Errorf("dynamic field key %q escapes the submission files directory", name)
	}
	return joined, nil
}

func (s *SubmissionService) SaveProject(submissionID string, project *bytes.Buffer) error {
	if submissionID == "" {
		return errEmptySubmissionID
	}
	s.logger.Infof("Saving project for submission %s", submissionID)
	projectPath := s.projectPath(submissionID)
	if err := s.fss.MkdirAll(projectPath); err != nil {
		s.logger.Errorf("Failed to create project directory for submission %s: %v", submissionID, err)
		return err
	}

	// Written before extraction, not after. Extraction is the step most
	// likely to fail, and working out why needs the archive that caused it —
	// storing it afterwards would keep evidence for exactly the submissions
	// that never needed investigating.
	if err := s.fss.WriteFile(s.archivePath(submissionID), project.Bytes()); err != nil {
		s.logger.Errorf("Failed to save uploaded archive for submission %s: %v", submissionID, err)
		return err
	}

	if err := s.zip.UnzipToPath(project, projectPath, true); err != nil {
		s.logger.Errorf("Failed to unzip project for submission %s: %v", submissionID, err)
		return err
	}

	s.logger.Infof("Successfully saved project for submission %s", submissionID)
	return nil
}

func (s *SubmissionService) CleanupSubmissionFiles(submissionID string) {
	if submissionID == "" {
		s.logger.Errorf("Refusing to cleanup submission files: empty submission ID would remove the whole submissions directory")
		return
	}
	s.logger.Infof("Cleaning up submission files for submission %s", submissionID)

	// Named explicitly rather than removing the submission directory
	// wholesale, so the uploaded archive survives. This runs on the failure
	// paths of submission creation — precisely the submissions worth
	// reproducing — and taking project.zip with it would destroy the only
	// evidence of what went wrong. A directory holding just project.zip
	// therefore reads as "this submission never landed".
	for _, path := range []string{
		s.projectPath(submissionID),
		s.filesPath(submissionID),
		s.metadataPath(submissionID),
	} {
		if err := s.fss.RemoveAll(path); err != nil {
			s.logger.Errorf("Failed to cleanup %s for submission %s: %v", path, submissionID, err)
		}
	}
}

func (s *SubmissionService) LoadMetadata(submissionID string) (map[string]string, error) {
	if submissionID == "" {
		return nil, errEmptySubmissionID
	}
	s.logger.Infof("Loading metadata for submission %s", submissionID)
	metadata, err := s.fss.ReadFile(s.metadataPath(submissionID))
	if err != nil {
		s.logger.Errorf("Failed to load metadata for submission %s: %v", submissionID, err)
		return nil, err
	}

	var metadataMap map[string]string
	if err := json.Unmarshal(metadata, &metadataMap); err != nil {
		s.logger.Errorf("Failed to unmarshal metadata for submission %s: %v", submissionID, err)
		return nil, err
	}

	return metadataMap, nil
}

func (s *SubmissionService) SaveDynamicFields(submissionID string, dynamicFields []*submissiondynamicfield.SubmissionDynamicField) error {
	if submissionID == "" {
		return errEmptySubmissionID
	}
	s.logger.Infof("Saving dynamic fields for submission %s", submissionID)
	filedir := s.filesPath(submissionID)
	if err := s.fss.MkdirAll(filedir); err != nil {
		s.logger.Errorf("Failed to create files directory for submission %s: %v", submissionID, err)
		return err
	}

	metadata := map[string]interface{}{}
	for _, df := range dynamicFields {
		if df.Type().IsText() {
			metadata[df.Key()] = df.Value()
		} else if df.Type().IsBool() {
			metadata[df.Key()] = df.Value()
		} else if df.Type().IsFile() {
			// The field key is project-defined and reaches here as a filename.
			// Contain it against filedir so a key such as "../../config.yaml"
			// from a malicious project bundle cannot write outside the
			// submission's files directory, regardless of upstream validation.
			target, err := containedPath(filedir, df.Key())
			if err != nil {
				s.logger.Errorf("Refusing unsafe dynamic field key %q for submission %s: %v", df.Key(), submissionID, err)
				return err
			}
			if err := s.fss.WriteFile(target, df.Value().([]byte)); err != nil {
				s.logger.Errorf("Failed to write file %s for submission %s: %v", df.Key(), submissionID, err)
				return err
			}
		}
	}

	jsonMetadata, err := json.Marshal(metadata)
	if err != nil {
		s.logger.Errorf("Failed to marshal metadata for submission %s: %v", submissionID, err)
		return err
	}

	if err := s.fss.WriteFile(s.metadataPath(submissionID), jsonMetadata); err != nil {
		s.logger.Errorf("Failed to write metadata file for submission %s: %v", submissionID, err)
		return err
	}

	s.logger.Infof("Successfully saved dynamic fields for submission %s", submissionID)
	return nil
}
