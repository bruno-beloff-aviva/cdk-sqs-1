package sub

import (
	"context"
	"encoding/json"
	"sqstest/manager/dbmanager"
	"sqstest/service/testmessage"
	"sqstest/service/testreception"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type ContinuousService struct {
	logger    *zapray.Logger
	dbManager dbmanager.DynamoManager
	id        string
}

func NewContinuousService(logger *zapray.Logger, cfg aws.Config, dbManager dbmanager.DynamoManager, id string) ContinuousService {
	return ContinuousService{logger: logger, dbManager: dbManager, id: id}
}

func (m ContinuousService) Receive(ctx context.Context, record events.SQSMessage) (err error) {
	m.logger.Debug("Receive", zap.String("record body", record.Body))

	var message testmessage.TestMessage
	var reception testreception.TestReception

	err = json.Unmarshal([]byte(record.Body), &message)
	if err != nil {
		return err
	}

	m.logger.Debug("Receive: ", zap.Any("message", message))

	// dbManager.Put...
	reception = testreception.NewTestReception(m.id, message)
	m.logger.Info("Receive: ", zap.Any("reception", reception))

	err = m.dbManager.Put(ctx, &reception)
	if err != nil {
		m.logger.Error("Receive: ", zap.Error(err))
	}

	return nil
}
