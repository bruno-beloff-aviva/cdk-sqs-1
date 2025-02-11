package service

import (
	"context"
	"encoding/json"
	"sqstest/service/testmessage"
	"sqstest/sqsmanager"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type PublishService struct {
	logger     *zapray.Logger
	sqsManager sqsmanager.SQSManager
	queueUrl   string
}

func NewPublishService(logger *zapray.Logger, cfg aws.Config, queueUrl string) PublishService {
	sqsManager := sqsmanager.NewSQSManager(logger, cfg)
	return PublishService{logger: logger, sqsManager: sqsManager, queueUrl: queueUrl}
}

func (m PublishService) Publish(ctx context.Context, clientId string, path string) (string, error) {
	m.logger.Info("Publish", zap.String("clientId", clientId))

	message := testmessage.NewTestMessage(clientId, path)

	jmsg, err := json.Marshal(message)
	strmsg := string(jmsg)

	if err != nil {
		panic(err)
	}

	return strmsg, m.sqsManager.Publish(ctx, m.queueUrl, strmsg)
}
