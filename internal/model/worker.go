package model

type EmailPayload struct {
	Email   string `json:"email"`
	Message string `json:"message"`
	Subject string `json:"subject"`
}

type SMSPayload struct {
	PhoneNumbers []string `json:"phone_numbers"`
	Message      string   `json:"message"`
}
