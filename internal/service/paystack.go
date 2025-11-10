package service

import (
	"errors"
	"konnect/internal/config"
	"konnect/internal/logger"

	"go.uber.org/zap"

	"github.com/go-resty/resty/v2"
)

type PaystackService struct {
	cfg        *config.Config
	logger     *logger.Logger
	httpClient *resty.Client
}

var (
	ErrPaystackServer = errors.New("paystack server error")
	ErrPaystackClient = errors.New("paystack client error")
)

// NewPaystackService creates a new instance of the Paystack service with a base http client.
// The http client must be closed on shutdown
func NewPaystackService(cfg *config.Config, logger *logger.Logger) *PaystackService {
	// base client with auth header
	httpClient := resty.New().SetBaseURL("https://api.paystack.co/")
	// httpClient.SetHeader("Authorization", "Bearer "+cfg.PaystackSecret)

	return &PaystackService{
		cfg:        cfg,
		httpClient: httpClient,
		logger:     logger,
	}
}

func (s *PaystackService) logPaystackError(err error, msg string, fields ...zap.Field) {
	s.logger.Error(msg, append(fields, zap.String("component", "paystack_service"), zap.Error(err))...)
}

// func (s *PaystackService) InitiateMobileMoneyCharge(amount int64, email string, phone string, provider model.PaystackProviderCode) (*model.PaystackCreateChargeResponse, error) {
// 	reference := s.GenerateReference()
// 	body := model.PaystackCreateCharge{
// 		Amount:    int(100 * amount),
// 		Email:     email,
// 		Reference: &reference,
// 		Currency:  "GHS",
// 	}
// 	body.MobileMoney.Phone = phone
// 	body.MobileMoney.Provider = provider

// 	var (
// 		resp    model.PaystackCreateChargeResponse
// 		errResp model.PaystackErrorResponse
// 	)

// 	res, err := s.httpClient.R().
// 		SetBody(body).
// 		SetResult(&resp).
// 		SetError(&errResp).
// 		Post("charge")

// 	if err != nil {
// 		s.logPaystackError(err, "http request failed for mobile money charge",
// 			zap.String("email", email),
// 			zap.String("phone", phone),
// 			zap.Int64("amount", amount),
// 			zap.String("provider", string(provider)),
// 			zap.String("reference", reference),
// 		)
// 		return nil, fmt.Errorf("%w: %v", ErrPaystackClient, err)
// 	}
// 	if res.IsError() {
// 		return nil, fmt.Errorf("%v", errResp.Message)
// 	}

// 	return &resp, nil
// }

// func (s *PaystackService) SubmitOTP(reference string, otp string) (*model.PaystackSubmitOtpResponse, error) {
// 	body := map[string]string{
// 		"otp":       otp,
// 		"reference": reference,
// 	}

// 	var (
// 		resp    model.PaystackSubmitOtpResponse
// 		errResp model.PaystackErrorResponse
// 	)

// 	res, err := s.httpClient.R().
// 		SetBody(body).
// 		SetResult(&resp).
// 		SetError(&errResp).
// 		Post("charge/submit_otp")

// 	if err != nil {
// 		s.logPaystackError(err, "http request failed for OTP submission", zap.String("reference", reference))
// 		return nil, fmt.Errorf("%w: %v", ErrPaystackServer, err)
// 	}

// 	if res.IsError() {
// 		return nil, fmt.Errorf("%v", errResp.Message)
// 	}

// 	return &resp, nil
// }

// func (s *PaystackService) InitiateTransaction(amount int64, email string) (*model.PaystackInitTransferResponse, error) {
// 	reference := s.GenerateReference()
// 	body := model.PaystackInitTransfer{
// 		Amount:    strconv.FormatInt(100*amount, 10),
// 		Email:     email,
// 		Reference: &reference,
// 	}

// 	var (
// 		resp    model.PaystackInitTransferResponse
// 		errResp model.PaystackErrorResponse
// 	)

// 	res, err := s.httpClient.R().
// 		SetBody(body).
// 		SetResult(&resp).
// 		SetError(&errResp).
// 		Post("transaction/initialize")

// 	if err != nil {
// 		s.logPaystackError(err, "http request failed for transaction initialization",
// 			zap.String("email", email),
// 			zap.Int64("amount", amount),
// 			zap.String("reference", reference),
// 		)
// 		return nil, fmt.Errorf("%w: %v", ErrPaystackServer, err)
// 	}
// 	if res.IsError() {
// 		return nil, fmt.Errorf("%v", errResp.Message)
// 	}

// 	return &resp, nil
// }

// func (s *PaystackService) VerifyTransaction(reference string) (*model.PaystackVerifyResponse, error) {
// 	var (
// 		resp    model.PaystackVerifyResponse
// 		errResp model.PaystackErrorResponse
// 	)

// 	res, err := s.httpClient.R().
// 		SetResult(&resp).
// 		SetError(&errResp).
// 		Get("transaction/verify/" + reference)

// 	if err != nil {
// 		s.logPaystackError(err, "http request failed for transaction verification", zap.String("reference", reference))
// 		return nil, fmt.Errorf("%w: %v", ErrPaystackServer, err)
// 	}
// 	if res.IsError() {
// 		return nil, fmt.Errorf("%v", errResp.Message)
// 	}

// 	return &resp, nil
// }

// // VerifyWebhookSignature validates the payload hash against the provided signature from the webhook headers
// func (s *PaystackService) VerifyWebhookSignature(payload []byte, signature string) bool {
// 	// hash of payload with paystack secret must match signature
// 	mac := hmac.New(sha512.New, []byte(s.cfg.PaystackSecret))
// 	mac.Write(payload)
// 	expectedMAC := mac.Sum(nil)
// 	expectedSignature := hex.EncodeToString(expectedMAC)

// 	return hmac.Equal([]byte(signature), []byte(expectedSignature))
// }

// // IsValidOrigin checks if the provided source IP matches paystack's provided IPs
// func (s *PaystackService) IsValidOrigin(ip string) bool {
// 	return slices.Contains(s.cfg.PaystackOrigins, ip)
// }

// // GenerateReference generates a random reference for the transaction
// func (s *PaystackService) GenerateReference() string {
// 	return "charge_" + strings.ReplaceAll(uuid.New().String(), "-", "")
// }

// // GetProvider retrieves the provider based on the phone number prefix
// func (s *PaystackService) GetProvider(phone string) (model.PaystackProviderCode, error) {
// 	// remove space and leading zero or 233
// 	phone = strings.TrimSpace(phone)
// 	phone = strings.ReplaceAll(phone, " ", "")
// 	phone = strings.ReplaceAll(phone, "+", "")

// 	if strings.HasPrefix(phone, "233") {
// 		phone = strings.TrimPrefix(phone, "233")
// 	} else if strings.HasPrefix(phone, "0") {
// 		phone = strings.TrimPrefix(phone, "0")
// 	}

// 	if len(phone) < 2 {
// 		return "", errors.New("invalid phone number")
// 	}

// 	switch phone[:2] {
// 	case "24", "54", "55", "59", "25":
// 		return model.MTN, nil
// 	case "20", "50":
// 		return model.Vodafone, nil
// 	case "26", "27", "56":
// 		return model.AirtelTigo, nil
// 	default:
// 		return "", errors.New("unknown service provider")
// 	}
// }
