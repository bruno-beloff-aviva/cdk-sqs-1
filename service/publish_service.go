package service

import (
	"context"
	"sqstest/sqsmanager"

	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type PublishService struct {
	logger     *zapray.Logger
	sqsManager sqsmanager.SQSManager
	queueUrl   string
}

func NewPublishService(logger *zapray.Logger, sqsManager sqsmanager.SQSManager, queueUrl string) PublishService {
	return PublishService{logger: logger, sqsManager: sqsManager, queueUrl: queueUrl}
}

func (m PublishService) Publish(ctx context.Context, clientId string) (message string, err error) {
	m.logger.Info("Publish", zap.String("clientId", clientId))
	message = "Message from " + clientId // TDOO: add datetime to the message

	return message, m.sqsManager.Publish(ctx, m.queueUrl, message)
}
