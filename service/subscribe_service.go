package service

import (
	"sqstest/sqsmanager"

	"github.com/aws/aws-lambda-go/events"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type SubscribeService struct {
	logger     *zapray.Logger
	sqsManager sqsmanager.SQSManager
	queueUrl   string
}

func NewSubscribeService(logger *zapray.Logger, sqsManager sqsmanager.SQSManager, queueUrl string) SubscribeService {
	return SubscribeService{logger: logger, sqsManager: sqsManager, queueUrl: queueUrl}
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
