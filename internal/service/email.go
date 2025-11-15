package service

import (
	"errors"
	"fmt"
	"konnect/internal/config"

	"github.com/go-resty/resty/v2"
)

type EmailService struct {
	cfg        *config.Config
	httpClient *resty.Client
}

var (
	ErrCourierEmailServer = errors.New("courier email server error")
)

func NewEmailService(cfg *config.Config) *EmailService {
	// base client with auth header
	httpClient := resty.New().SetBaseURL("https://api.courier.com")
	httpClient.SetHeader("Authorization", "Bearer "+cfg.CourierAPIKey)

	return &EmailService{
		cfg:        cfg,
		httpClient: httpClient,
	}
}

func (s *EmailService) Send(email string, message string, subject string) error {
	// courier email without template payload
	body := map[string]any{
		"message": map[string]any{
			"to": map[string]string{
				"email": email,
			},
			"content": map[string]string{
				"title": subject,
				"body":  message,
			},
		},
	}

	res, err := s.httpClient.R().
		SetBody(body).
		Post("/send")

	// server error. request possibly hanging
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCourierEmailServer, err)
	}
	if res.IsError() {
		return fmt.Errorf("%v", res.Error())
	}
	return nil
}
