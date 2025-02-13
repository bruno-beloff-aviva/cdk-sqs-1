package pub

import (
	"context"
	"encoding/json"
	"sqstest/manager/snsmanager"
	"sqstest/service/testmessage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type SNSPubService struct {
	logger     *zapray.Logger
	snsManager snsmanager.SNSManager
	topicArn   string
}

func NewSNSPubService(logger *zapray.Logger, cfg aws.Config, topicArn string) SNSPubService {
	snsManager := snsmanager.NewSNSManager(logger, cfg)

	return SNSPubService{logger: logger, snsManager: snsManager, topicArn: topicArn}
}

func (m SNSPubService) Publish(ctx context.Context, clientId string, path string) (testmessage.TestMessage, error) {
	m.logger.Debug("Publish", zap.String("clientId", clientId))

	message := testmessage.NewTestMessage(clientId, path)

	jmsg, err := json.Marshal(message)
	strmsg := string(jmsg)

	if err != nil {
		panic(err)
	}

	m.logger.Info("Publish", zap.Any("message", message))

	return message, m.snsManager.Pub(ctx, m.topicArn, strmsg)
}
