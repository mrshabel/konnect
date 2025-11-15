package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"konnect/internal/model"
	"log"

	"github.com/hibiken/asynq"
)

// unique job type for the email sending job
const (
	TypeEmailDelivery = "email:delivery"
)

// the email sender
type EmailDispatcher interface {
	Send(email string, message string, subject string) error
}

// NewEmailDeliveryJob creates an email dispatch job
func NewEmailDeliveryJob(client *asynq.Client, data model.EmailPayload) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeEmailDelivery, payload)
	info, err := client.Enqueue(task, asynq.Queue("email"), asynq.MaxRetry(5))
	if err != nil {
		return err
	}
	log.Printf("enqueued email job: id=%s queue=%s\n", info.ID, info.Queue)
	return nil
}

// EmailProcessor implements asynq.Handler interface
type EmailProcessor struct {
	Dispatcher EmailDispatcher
}

func (p *EmailProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload model.EmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal sms payload: %v: %w", err, asynq.SkipRetry)
	}

	// dispatch email
	return p.Dispatcher.Send(payload.Email, payload.Message, payload.Subject)
}

func NewEmailProcessor(dispatcher EmailDispatcher) *EmailProcessor {
	return &EmailProcessor{
		Dispatcher: dispatcher,
	}
}
