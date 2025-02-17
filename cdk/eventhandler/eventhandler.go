package eventhandler

import (
	"sqstest/aviva/sqs"

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

// common to all event handlers
type EventHandlerProps struct {
	SubscriptionTopic   awssns.Topic
	QueueKey            awskms.IKey
	QueueMaxRetries     int
	MessageTable        awsdynamodb.ITable
	CloudwatchDashboard awscloudwatch.Dashboard
}

// specific to an event handler
type EventHandler struct {
	EventName   string
	QueueName   string
	HandlerId   string
	Entry       string
	Environment map[string]*string
}

func (e EventHandler) Setup(stack awscdk.Stack, props EventHandlerProps) {
	queue := e.setupQueue(stack, props)

	subProps := awssnssubscriptions.SqsSubscriptionProps{
		RawMessageDelivery: aws.Bool(true),
	}
	props.SubscriptionTopic.AddSubscription(awssnssubscriptions.NewSqsSubscription(queue, &subProps))

	if e.HandlerId == "" {
		return
	}

	handler := e.setupSubHandler(stack, queue)
	queue.GrantConsumeMessages(handler)
	props.MessageTable.GrantReadWriteData(handler)
}

func (e EventHandler) setupQueue(stack awscdk.Stack, props EventHandlerProps) awssqs.IQueue {
	queueProps := sqs.SqsQueueWithDLQProps{
		Stack:                    stack,
		QueueName:                e.QueueName,
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

func (e EventHandler) setupSubHandler(stack awscdk.Stack, queue awssqs.IQueue) awslambdago.GoFunction {
	handlerProps := awslambdago.GoFunctionProps{
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String(e.Entry),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(28)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Tracing:       awslambda.Tracing_ACTIVE,
		Environment:   &e.Environment,
	}

	handler := awslambdago.NewGoFunction(stack, aws.String(e.HandlerId), &handlerProps)

	eventSourceProps := awslambdaeventsources.SqsEventSourceProps{}
	handler.AddEventSource(awslambdaeventsources.NewSqsEventSource(queue, &eventSourceProps))

	return handler
}

// func (e *EventHandler) AddCloudwatchDashboardMetrics(region string, props EventHandlerProps, handler awslambdago.GoFunction) {
// 	invocationsMetric := e.CreateLambdaMetric(region, "Invocations", handler.FunctionName(), "Sum")
// 	errorsMetric := e.CreateLambdaMetric(region, "Errors", handler.FunctionName(), "Sum")

// 	invocationsAndErrors := e.CreateGraphWidget(region, fmt.Sprintf("%s Invocations and Errors", e.eventName), []awscloudwatch.IMetric{invocationsMetric, errorsMetric})

// 	row := awscloudwatch.NewRow(invocationsAndErrors)
// 	props.CloudwatchDashboard.AddWidgets(row)
// }

// func (e *EventHandler) CreateLambdaMetric(region string, metricName string, functionName *string, statistic string) awscloudwatch.IMetric {
// 	return awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
// 		Region:     jsii.String(region),
// 		Namespace:  jsii.String("AWS/Lambda"),
// 		MetricName: jsii.String(metricName),
// 		DimensionsMap: &map[string]*string{
// 			"FunctionName": functionName,
// 		},
// 		Period:    awscdk.Duration_Minutes(jsii.Number(5)),
// 		Statistic: jsii.String(statistic),
// 	})
// }

// func (e *EventHandler) CreateCustomMetric(region string, namespace, metricName, eventName, statistic string) awscloudwatch.IMetric {
// 	return awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
// 		Region:     jsii.String(region),
// 		Namespace:  jsii.String(namespace),
// 		MetricName: jsii.String(metricName),
// 		DimensionsMap: &map[string]*string{
// 			"event": jsii.String(eventName),
// 		},
// 		Period:    awscdk.Duration_Minutes(jsii.Number(5)),
// 		Statistic: jsii.String(statistic),
// 	})
// }

// func (e *EventHandler) CreateGraphWidget(region string, title string, metrics []awscloudwatch.IMetric) awscloudwatch.GraphWidget {
// 	return awscloudwatch.NewGraphWidget(&awscloudwatch.GraphWidgetProps{
// 		Region: jsii.String(region),
// 		Title:  jsii.String(title),
// 		Left:   &metrics,
// 		Height: jsii.Number(6),
// 		Width:  jsii.Number(6),
// 	})
// }

// func (e *EventHandler) CreateSingleValueWidget(region string, title string, metrics []awscloudwatch.IMetric) awscloudwatch.SingleValueWidget {
// 	return awscloudwatch.NewSingleValueWidget(&awscloudwatch.SingleValueWidgetProps{
// 		Region:               jsii.String(region),
// 		Title:                jsii.String(title),
// 		Metrics:              &metrics,
// 		SetPeriodToTimeRange: jsii.Bool(true),
// 		Height:               jsii.Number(6),
// 		Width:                jsii.Number(4),
// 	})
// }
