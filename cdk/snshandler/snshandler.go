package snshandler

import (
	"fmt"
	"sqstest/aviva/sqs"
	"sqstest/cdk/dashboard"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
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

type SNSCommonProps struct {
	QueueKey        awskms.IKey
	QueueMaxRetries int
	MessageTable    awsdynamodb.ITable
	Dashboard       dashboard.Dashboard
}

type SNSBuilder struct {
	SubscriptionTopic awssns.Topic
	QueueName         string
	HandlerId         string
	Entry             string
	Environment       map[string]*string
}

type SNSConstruct struct {
	Builder   SNSBuilder
	Queue     awssqs.IQueue
	Handler   awslambdago.GoFunction
	Dashboard dashboard.Dashboard
}

func (h SNSBuilder) Setup(stack awscdk.Stack, props SNSCommonProps) SNSConstruct {
	var c SNSConstruct

	c.Builder = h
	c.Dashboard = props.Dashboard
	c.Queue = h.setupQueue(stack, props)

	subProps := awssnssubscriptions.SqsSubscriptionProps{
		RawMessageDelivery: aws.Bool(true),
	}
	h.SubscriptionTopic.AddSubscription(awssnssubscriptions.NewSqsSubscription(c.Queue, &subProps))

	if h.HandlerId == "" {
		return c
	}

	c.Handler = h.setupSubHandler(stack, c.Queue)
	c.Queue.GrantConsumeMessages(c.Handler)
	props.MessageTable.GrantReadWriteData(c.Handler)

	return c
}

func (h SNSBuilder) setupQueue(stack awscdk.Stack, props SNSCommonProps) awssqs.IQueue {
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

func (h SNSBuilder) setupSubHandler(stack awscdk.Stack, queue awssqs.IQueue) awslambdago.GoFunction {
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

func (c SNSConstruct) MetricsGraphWidget(stack awscdk.Stack) awscloudwatch.GraphWidget {
	region := *stack.Region()

	invocationsMetric := c.Dashboard.CreateLambdaMetric(region, "Invocations", c.Handler.FunctionName(), "Sum")
	errorsMetric := c.Dashboard.CreateLambdaMetric(region, "Errors", c.Handler.FunctionName(), "Sum")
	metrics := []awscloudwatch.IMetric{invocationsMetric, errorsMetric}

	return c.Dashboard.CreateGraphWidget(region, fmt.Sprintf("%s - Invocations and Errors", c.Builder.HandlerId), metrics)
}
