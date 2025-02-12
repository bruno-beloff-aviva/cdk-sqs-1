package service

import (
	"encoding/json"
	"errors"
	"sqstest/service/testmessage"
	"sqstest/sqsmanager"
	"strings"
	"time"

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

func (m SubscribeService) Receive(record events.SQSMessage) (err error) {
	m.logger.Debug("Receive", zap.String("record body", record.Body))
	var message testmessage.TestMessage

	err = json.Unmarshal([]byte(record.Body), &message)
	if err != nil {
		return err
	}

	m.logger.Info("Receive: ", zap.Any("message", message))

	// sleep...
	if strings.Contains(message.Path, "sleep") {
		m.logger.Info("*** sleep")
		time.Sleep(10 * time.Second)
	}

	// error...
	if strings.Contains(message.Path, "error") {
		m.logger.Info("*** error")
		return errors.New(message.Path)
	}

	// panic...
	if strings.Contains(message.Path, "panic") {
		m.logger.Info("*** panic")
		panic(message.Path)
	}

	// TODO: put in dynamodb

	return nil
}
