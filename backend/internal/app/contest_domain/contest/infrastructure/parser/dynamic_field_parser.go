package parser

import (
	"fmt"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/submissiondynamicfield"

	"gorm.io/gorm"
)

type DynamicFieldParser interface {
	Parse(contestProjectID int32, dynamicFields map[string]string) ([]*submissiondynamicfield.SubmissionDynamicField, error)
}

type dynamicFieldParser struct {
	db     *gorm.DB
	logger sharedtypes.Logger
}

func NewDynamicFieldParser(
	db *gorm.DB,
	logger sharedtypes.Logger,
) DynamicFieldParser {
	return &dynamicFieldParser{
		db:     db,
		logger: logger,
	}
}

func (p *dynamicFieldParser) Parse(
	contestProjectID int32,
	dynamicFields map[string]string,
) ([]*submissiondynamicfield.SubmissionDynamicField, error) {
	p.logger.Debugf("Starting to parse dynamic fields", "contestProjectId", contestProjectID)

	cp := &model.ContestProject{}
	if err := p.db.First(cp, contestProjectID).Error; err != nil {
		p.logger.Errorf("Failed to find contest project", "contestProjectId", contestProjectID, "error", err)
		return nil, err
	}

	dfs := []*model.DynamicField{}
	if err := p.db.Where("project_id = ?", cp.ProjectID).Find(&dfs).Error; err != nil {
		p.logger.Errorf("Failed to find dynamic fields", "projectId", cp.ProjectID, "error", err)
		return nil, err
	}

	parsedFields := []*submissiondynamicfield.SubmissionDynamicField{}
	for _, df := range dfs {
		val, ok := dynamicFields[df.Name]
		if !ok && df.IsRequired {
			p.logger.Debugf("Required dynamic field not found", "fieldName", df.Name)
			return nil, fmt.Errorf("dynamic field \"%s\" is required", df.Name)
		}
		if !ok && df.DefaultValue != nil {
			val = *df.DefaultValue
		}
		if !ok {
			continue
		}

		parsedField, err := p.convertToSubmissionDynamicField(df, val)
		if err != nil {
			p.logger.Errorf("Failed to convert dynamic field value", "fieldName", df.Name, "error", err)
			return nil, err
		}

		parsedFields = append(parsedFields, parsedField)
	}

	p.logger.Debugf("Successfully parsed dynamic fields", "contestProjectId", contestProjectID, "fieldCount", len(parsedFields))
	return parsedFields, nil
}

func (p *dynamicFieldParser) convertToSubmissionDynamicField(df *model.DynamicField, value string) (*submissiondynamicfield.SubmissionDynamicField, error) {
	switch df.Type {
	case "file":
		return submissiondynamicfield.NewFile(df.Name, value)
	case "text":
		return submissiondynamicfield.NewText(df.Name, value), nil
	case "bool":
		return submissiondynamicfield.NewBool(df.Name, value)
	default:
		p.logger.Warnf("Unknown dynamic field type, falling back to text", "fieldName", df.Name, "type", df.Type)
		return submissiondynamicfield.NewText(df.Name, value), nil
	}
}
