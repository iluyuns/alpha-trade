package email

import "context"

type EmailService interface {
	SendEmail(ctx context.Context, senderEmail string, to []string, subject string, body string) (messageId string, err error)
}
