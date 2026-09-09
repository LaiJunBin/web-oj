package clarification

import (
	"fmt"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type Clarification struct {
	id               int32
	contestProjectID int32
	subject          Subject
	content          string
	senderUserID     *int32
	receiverUserID   *int32
	sendToAll        bool
	isAnswered       bool
	replyToID        *int32

	sharedtypes.TimeStamps
}

type ClarificationConfig struct {
	ContestProjectID int32
	Subject          string
	Content          string
	SenderUserID     *int32
	ReceiverUserID   *int32
	SendToAll        bool
	IsAnswered       bool
	ReplyToID        *int32
}

func New(config *ClarificationConfig) (*Clarification, error) {
	subject, err := NewSubject(config.Subject)
	if err != nil {
		return nil, err
	}

	return &Clarification{
		contestProjectID: config.ContestProjectID,
		subject:          subject,
		content:          config.Content,
		senderUserID:     config.SenderUserID,
		receiverUserID:   config.ReceiverUserID,
		sendToAll:        config.SendToAll,
		isAnswered:       config.IsAnswered,
		replyToID:        config.ReplyToID,
	}, nil
}

func (c *Clarification) ID() int32 {
	return c.id
}

func (c *Clarification) ContestProjectID() int32 {
	return c.contestProjectID
}

func (c *Clarification) Subject() Subject {
	return c.subject
}

func (c *Clarification) Content() string {
	return c.content
}

func (c *Clarification) SenderUserID() *int32 {
	return c.senderUserID
}

func (c *Clarification) ReceiverUserID() *int32 {
	return c.receiverUserID
}

func (c *Clarification) SendToAll() bool {
	return c.sendToAll
}

func (c *Clarification) ReplyToID() *int32 {
	return c.replyToID
}

func (c *Clarification) IsAnswered() bool {
	return c.isAnswered
}

func (c *Clarification) SetID(id int32) {
	c.id = id
}

func (c *Clarification) SetIsAnswered(isAnswered bool) {
	c.isAnswered = isAnswered
}

type ReplyConfig struct {
	SenderUserID int32
	Subject      *string
	Content      string
	SendToAll    bool
}

func (c *Clarification) Reply(config *ReplyConfig) (*Clarification, error) {
	subject := fmt.Sprintf("Re: %s", c.subject.String())
	if config.Subject != nil {
		subject = *config.Subject
	}

	return New(&ClarificationConfig{
		ContestProjectID: c.contestProjectID,
		Subject:          subject,
		Content:          config.Content,
		SenderUserID:     &config.SenderUserID,
		ReceiverUserID:   c.senderUserID,
		SendToAll:        config.SendToAll,
		ReplyToID:        &c.id,
	})
}
