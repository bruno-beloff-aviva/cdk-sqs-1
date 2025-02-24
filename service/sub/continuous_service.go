package sub

import (
	"context"
	"encoding/json"
	"sqstest/lambda/handler/singleshot"
	"sqstest/manager/dbmanager"
	"sqstest/service/testmessage"
	"sqstest/service/testreception"
	"sqstest/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type ContinuousService struct {
	singleshot.SingleShotService[testmessage.TestMessage]
	// singleshot.SingleShotService[testmessage.TestMessage]
	logger    *zapray.Logger
	dbManager dbmanager.DynamoManager
	id        string
	// gateway   singleshot.SingleshotGateway[testmessage.TestMessage]
}

func NewContinuousService(logger *zapray.Logger, cfg aws.Config, dbManager dbmanager.DynamoManager, id string) ContinuousService {
	// service := ContinuousService{
	// 	singleshot.SingleShotService[testmessage.TestMessage]{
	// 		Gateway: singleshot.NewSingleshotGateway[testmessage.TestMessage](logger, services.NullEventHasBeenProcessed, services.NullMarkEventAsProcessed),
	// 	},
	// 	logger,
	// 	dbManager,
	// 	id,
	// }

	// 	m.gateway = NewSingleshotGateway[T](m.logger, m.(SingleshotHandler[T]), eventHasBeenProcessed, EventAsProcessed)

	service := ContinuousService{logger: logger, dbManager: dbManager, id: id}
	// service.attachGateway(services.NullEventHasBeenProcessed, services.NullMarkEventAsProcessed)

	service.NewGateway(logger, services.NullEventHasBeenProcessed, services.NullMarkEventAsProcessed)

	// service.gateway = singleshot.NewSingleshotGateway(logger, service, services.NullEventHasBeenProcessed, services.NullMarkEventAsProcessed)

	return service
}

// func (m *ContinuousService) attachGateway(eventHasBeenProcessed services.EventHasBeenProcessedFunc, EventAsProcessed services.MarkEventAsProcessedFunc) {
// 	m.gateway = singleshot.NewSingleshotGateway(m.logger, m, eventHasBeenProcessed, EventAsProcessed)
// }

func (m ContinuousService) Receive(ctx context.Context, record events.SQSMessage) (err error) {
	m.logger.Info("Receive", zap.String("record body", record.Body))

	var message testmessage.TestMessage

	err = json.Unmarshal([]byte(record.Body), &message)
	if err != nil {
		return err
	}

	err = m.Gateway.ProcessOnce(ctx, message)

	if err == nil {
		// TODO: success metrics here
	}

	return err
}

func (m ContinuousService) UniqueID(event testmessage.TestMessage) (policyOrQuoteID string, eventID string, err error) {
	return event.Client, event.Sent, nil
}

func (m ContinuousService) Process(ctx context.Context, event testmessage.TestMessage) (err error) {
	m.logger.Debug("Process: ", zap.Any("event", event))

	// dbManager.Put...
	reception := testreception.NewTestReception(m.id, event)
	m.logger.Info("Process: ", zap.Any("reception", reception))

	err = m.dbManager.Put(ctx, &reception)
	if err != nil {
		m.logger.Error("Put: ", zap.Error(err))
	}

	return err
}
