package snshandler

import (
	"sqstest/aviva/sqs"
	"sqstest/cdk/dashboard"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssnssubscriptions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	awslambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/aws-sdk-go/aws"
)

// common to all SNS handlers
type SNSHandlerProps struct {
	QueueKey            awskms.IKey
	QueueMaxRetries     int
	MessageTable        awsdynamodb.ITable
	CloudwatchDashboard dashboard.Dashboard
}

// specific to an SNS handler
type SNSHandler struct {
	SubscriptionTopic awssns.Topic
	QueueName         string
	HandlerId         string
	Entry             string
	Environment       map[string]*string
}

func (h SNSHandler) Setup(stack awscdk.Stack, props SNSHandlerProps) awslambdago.GoFunction {
	queue := h.setupQueue(stack, props)

	subProps := awssnssubscriptions.SqsSubscriptionProps{
		RawMessageDelivery: aws.Bool(true),
	}
	h.SubscriptionTopic.AddSubscription(awssnssubscriptions.NewSqsSubscription(queue, &subProps))

	if h.HandlerId == "" {
		return nil
	}

	handler := h.setupSubHandler(stack, queue)
	queue.GrantConsumeMessages(handler)
	props.MessageTable.GrantReadWriteData(handler)

	return handler
}

func (h SNSHandler) setupQueue(stack awscdk.Stack, props SNSHandlerProps) awssqs.IQueue {
	queueProps := sqs.SqsQueueWithDLQProps{
		Stack:                    stack,
		QueueName:                h.QueueName,
		SQSKey:                   props.QueueKey,
		QMaxReceiveCount:         props.QueueMaxRetries,
		QAlarmPeriod:             1,
		QAlarmThreshold:          1,
		QAlarmEvaluationPeriod:   1,
		DLQAlarmPeriod:           1,
		DLQAlarmThreshold:        1,
		DLQAlarmEvaluationPeriod: 1,
	}

	return sqs.NewSqsQueueWithDLQ(queueProps)
}

func (h SNSHandler) setupSubHandler(stack awscdk.Stack, queue awssqs.IQueue) awslambdago.GoFunction {
	handlerProps := awslambdago.GoFunctionProps{
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String(h.Entry),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(28)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Tracing:       awslambda.Tracing_ACTIVE,
		Environment:   &h.Environment,
	}

	handler := awslambdago.NewGoFunction(stack, aws.String(h.HandlerId), &handlerProps)

	eventSourceProps := awslambdaeventsources.SqsEventSourceProps{}
	handler.AddEventSource(awslambdaeventsources.NewSqsEventSource(queue, &eventSourceProps))

	return handler
}
