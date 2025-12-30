package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/iluyuns/alpha-trade/internal/config"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

type awsSES struct {
	client *ses.Client
	cfg    *config.AWSConfig
}

func NewAWSSES(cfg *config.AWSConfig) EmailService {
	awsCfg := aws.Config{
		Region: cfg.SESConfig.Region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	}
	// 开启 AWS SDK 自动追踪，这样你就能在 Jaeger 里看到 SES.SendEmail 的耗时
	otelaws.AppendMiddlewares(&awsCfg.APIOptions)

	return &awsSES{
		client: ses.NewFromConfig(awsCfg),
		cfg:    cfg,
	}
}

func (s *awsSES) SendEmail(ctx context.Context, senderEmail string, to []string, subject string, body string) (messageId string, err error) {
	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: to,
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Data: &body,
				},
			},
			Subject: &types.Content{
				Data: &subject,
			},
		},
		Source: &senderEmail,
	}
	output, err := s.client.SendEmail(ctx, input)
	if err != nil {
		return "", fmt.Errorf("ses: send email failed: %w", err)
	}

	return *output.MessageId, nil
}
