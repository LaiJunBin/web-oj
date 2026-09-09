package service

import (
	"bytes"
	"judge-backend/internal/app/contest_domain/contest/domain/service"
	"judge-backend/internal/app/contest_domain/contest/infrastructure/parser"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type ProjectSubmitPublisher struct {
	eventbus sharedtypes.EventBus
	parser   parser.DynamicFieldParser
	logger   sharedtypes.Logger
}

func NewProjectSubmitPublisher(
	eventbus sharedtypes.EventBus,
	parser parser.DynamicFieldParser,
	logger sharedtypes.Logger,
) service.ProjectSubmitPublisher {
	return &ProjectSubmitPublisher{
		eventbus: eventbus,
		parser:   parser,
		logger:   logger,
	}
}

func (p *ProjectSubmitPublisher) Publish(
	userID int32,
	contestProjectID int32,
	isPartial bool,
	gradingItemIDs []int32,
	project *bytes.Buffer,
	dynamicFields map[string]string,
) error {
	p.logger.Debugf("Starting to publish project submitted event", "contestProjectId", contestProjectID, "userId", userID)

	parsedFields, err := p.parser.Parse(contestProjectID, dynamicFields)
	if err != nil {
		// Surface it: the fields are part of the submission, and a parse failure
		// used to be swallowed here, answering the contestant 200 while the
		// submission silently never happened.
		p.logger.Errorf("Failed to parse dynamic fields", "contestProjectId", contestProjectID, "error", err)
		return err
	}

	event := sharedevent.NewProjectSubmitted(userID, contestProjectID, isPartial, gradingItemIDs, project, parsedFields)
	p.eventbus.Publish(event)

	p.logger.Debugf("Successfully published project submitted event", "contestProjectId", contestProjectID, "userId", userID)
	return nil
}
