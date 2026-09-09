package clarification

import (
	"fmt"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

type Clarification struct {
	id                   int32
	contestProjectID     int32
	sendUserID           int32
	receiverUserID       int32
	isAdminBroadcast     bool
	subject              Subject
	content              string
	replyClarificationID *int32
}

type ClarificationConfig struct {
	ContestProjectID     int32
	SenderUserID         int32
	ReceiverUserID       int32
	IsAdminBroadcast     bool
	Subject              string
	Content              string
	ReplyClarificationID *int32
}

func New(config *ClarificationConfig) (*Clarification, error) {
	subject, err := NewSubject(config.Subject)
	if err != nil {
		return nil, httperror.New(http.StatusBadRequest, "invalid clarification subject")
	}

	return &Clarification{
		contestProjectID:     config.ContestProjectID,
		sendUserID:           config.SenderUserID,
		receiverUserID:       config.ReceiverUserID,
		isAdminBroadcast:     config.IsAdminBroadcast,
		subject:              subject,
		content:              config.Content,
		replyClarificationID: config.ReplyClarificationID,
	}, nil
}

func (c *Clarification) ID() int32 {
	return c.id
}

func (c *Clarification) ContestProjectID() int32 {
	return c.contestProjectID
}

func (c *Clarification) SenderUserID() int32 {
	return c.sendUserID
}

func (c *Clarification) ReceiverUserID() int32 {
	return c.receiverUserID
}

func (c *Clarification) IsAdminBroadcast() bool {
	return c.isAdminBroadcast
}

func (c *Clarification) Subject() Subject {
	return c.subject
}

func (c *Clarification) Content() string {
	return c.content
}

func (c *Clarification) ReplyClarificationID() *int32 {
	return c.replyClarificationID
}

func (c *Clarification) SetID(id int32) {
	c.id = id
}

func (c *Clarification) CanReply(UserID int32) bool {
	return c.isAdminBroadcast || c.receiverUserID == UserID || c.sendUserID == UserID
}

type ReplyConfig struct {
	SendUserID int32
	Content    string
}

func (c *Clarification) Reply(config *ReplyConfig) (*Clarification, error) {
	if config.Content == "" {
		return nil, httperror.New(http.StatusBadRequest, "reply content is required")
	}

	return New(&ClarificationConfig{
		SenderUserID:         config.SendUserID,
		ContestProjectID:     c.ContestProjectID(),
		Subject:              fmt.Sprintf("Re: %s", c.Subject().String()),
		Content:              config.Content,
		ReplyClarificationID: &c.id,
	})
}
