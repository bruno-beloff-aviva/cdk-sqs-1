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

type PubService struct {
	logger     *zapray.Logger
	sqsManager sqsmanager.SQSManager
	queueUrl   string
}

func NewPubService(logger *zapray.Logger, cfg aws.Config, queueUrl string) PubService {
	sqsManager := sqsmanager.NewSQSManager(logger, cfg)

	return PubService{logger: logger, sqsManager: sqsManager, queueUrl: queueUrl}
}

func (m PubService) Publish(ctx context.Context, clientId string, path string) (testmessage.TestMessage, error) {
	m.logger.Debug("Publish", zap.String("clientId", clientId))

	message := testmessage.NewTestMessage(clientId, path)

	jmsg, err := json.Marshal(message)
	strmsg := string(jmsg)

	if err != nil {
		panic(err)
	}

	m.logger.Info("Publish", zap.Any("message", message))

	return message, m.sqsManager.Pub(ctx, m.queueUrl, strmsg)
}
