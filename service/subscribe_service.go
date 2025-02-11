package service

import (
	"sqstest/sqsmanager"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type SubscribeService struct {
	logger     *zapray.Logger
	sqsManager sqsmanager.SQSManager
}

func NewSubscribeService(logger *zapray.Logger, cfg aws.Config) SubscribeService {
	sqsManager := sqsmanager.NewSQSManager(logger, cfg)
	return SubscribeService{logger: logger, sqsManager: sqsManager}
}

// func  Publish(ctx context.Context, clientId string) (string, error) {
// 	message := testmessage.NewTestMessage(clientId)

// 	jmsg, err := json.Marshal(message)
// 	strmsg := string(jmsg)

// 	if err != nil {
// 		panic(err)
// 	}

// 	return strmsg, m.sqsManager.Publish(ctx, m.queueUrl, strmsg)
// }

func (m SubscribeService) Process(record events.SQSMessage) error {
	m.logger.Info("Process", zap.String("record body", record.Body))

	// TODO: Do interesting work based on the new message
	return nil
}
