package service

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type SubscriptionService interface {
	Receive(ctx context.Context, record events.SQSMessage) (err error)
}
