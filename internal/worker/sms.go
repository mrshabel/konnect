package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"konnect/internal/model"
	"log"

	"github.com/hibiken/asynq"
)

// unique task type for the sms sending job
const (
	TypeSMSDelivery = "sms:delivery"
)

// the sms sender
type SMSDispatcher interface {
	Send(phoneNumbers []string, message string) error
}

// NewSMSDeliveryJob creates an email dispatch job
func NewSMSDeliveryJob(client *asynq.Client, data model.SMSPayload) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeSMSDelivery, payload)
	info, err := client.Enqueue(task, asynq.Queue("sms"), asynq.MaxRetry(5))
	if err != nil {
		return err
	}
	log.Printf("enqueued sms job: id=%s queue=%s\n", info.ID, info.Queue)
	return nil
}

// SMSProcessor implements asynq.Handler interface
type SMSProcessor struct {
	Dispatcher SMSDispatcher
}

func (p *SMSProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload model.SMSPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal sms payload: %v: %w", err, asynq.SkipRetry)
	}

	// dispatch sms
	return p.Dispatcher.Send(payload.PhoneNumbers, payload.Message)
}

func NewSMSProcessor(dispatcher SMSDispatcher) *SMSProcessor {
	return &SMSProcessor{
		Dispatcher: dispatcher,
	}
}
