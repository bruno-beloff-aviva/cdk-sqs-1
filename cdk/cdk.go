package main

// https://github.com/aviva-verde/cdk-standards.git
// https://docs.aws.amazon.com/cdk/v2/guide/resources.html
// https://docs.aws.amazon.com/code-library/latest/ug/go_2_sqs_code_examples.html
// https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_lambda_event_sources.SqsEventSource.html

import (
	sqs "sqstest/aviva/sqs"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// TODO: build a dashboard

const project = "SQS1"
const version = "0.1.8"

const queue1Name = "TestQueue1"
const queue2Name = "TestQueue2"
const queue3Name = "TestQueue3"
const queueMaxRetries = 3

const tableName = "TestMessageTableV1"
const tableId = project + tableName

const topicName = "TestTopic"
const topicId = project + topicName

const pubHandlerId = project + "PubHandler"
const pubEndpointId = project + "PubEndpoint"

const continuousSubHandlerId = project + "ContinuousHandler"
const suspendableSubHandlerId = project + "SudspendableHandler"

const stackId = project + "Stack"

type CdkWorkshopStackProps struct { //	TODO: make use of this - make New... functions as methods?
	awscdk.StackProps
}

func NewMessageTable(stack awscdk.Stack, id string, name string) awsdynamodb.ITable {
	tableProps := awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{Name: aws.String("PK"), Type: awsdynamodb.AttributeType_STRING},
		SortKey:      &awsdynamodb.Attribute{Name: aws.String("Path"), Type: awsdynamodb.AttributeType_STRING},
		TableName:    aws.String(name),
	}

	return awsdynamodb.NewTable(stack, aws.String(id), &tableProps)
}

func NewTopic(stack awscdk.Stack, id string, name string) awssns.Topic {
	topicProps := awssns.TopicProps{
		DisplayName: aws.String(name),
	}

	return awssns.NewTopic(stack, aws.String(id), &topicProps)
}

func NewQueue(stack awscdk.Stack, name string) awssqs.IQueue {
	keyProps := awskms.KeyProps{
		Alias:             aws.String(name + "QueueKey"),
		EnableKeyRotation: aws.Bool(true),
	}

	queueKey := awskms.NewKey(stack, aws.String(name+"Key"), &keyProps)

	queueProps := sqs.SqsQueueWithDLQProps{
		Stack:                    stack,
		QueueName:                name,
		SQSKey:                   queueKey, //	TODO: attempt to remove this
		QMaxReceiveCount:         queueMaxRetries,
		QAlarmPeriod:             1,
		QAlarmThreshold:          1,
		QAlarmEvaluationPeriod:   1,
		DLQAlarmPeriod:           1,
		DLQAlarmThreshold:        1,
		DLQAlarmEvaluationPeriod: 1,
	}

	return sqs.NewSqsQueueWithDLQ(queueProps)
}

func NewPubHandler(stack awscdk.Stack, id string, topic awssns.Topic) awslambdago.GoFunction {
	lambdaEnv := map[string]*string{
		"VERSION":   aws.String(version),
		"TOPIC_ARN": topic.TopicArn(),
	}

	handlerProps := awslambdago.GoFunctionProps{
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String("lambda/pub/"),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(28)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Tracing:       awslambda.Tracing_ACTIVE,
		Environment:   &lambdaEnv,
	}

	return awslambdago.NewGoFunction(stack, aws.String(id), &handlerProps)
}

func NewContinuousSubHandler(stack awscdk.Stack, id string, queue awssqs.IQueue) awslambdago.GoFunction {
	lambdaEnv := map[string]*string{
		"VERSION":            aws.String(version),
		"MESSAGE_TABLE_NAME": aws.String(tableName),
	}

	handlerProps := awslambdago.GoFunctionProps{
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String("lambda/subcontinuous/"),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(28)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Tracing:       awslambda.Tracing_ACTIVE,
		Environment:   &lambdaEnv,
	}

	handler := awslambdago.NewGoFunction(stack, aws.String(id), &handlerProps)

	eventSourceProps := awslambdaeventsources.SqsEventSourceProps{}
	handler.AddEventSource(awslambdaeventsources.NewSqsEventSource(queue, &eventSourceProps))

	return handler
}

func NewSuspendableSubHandler(stack awscdk.Stack, id string, queue awssqs.IQueue) awslambdago.GoFunction {
	lambdaEnv := map[string]*string{
		"VERSION":            aws.String(version),
		"MESSAGE_TABLE_NAME": aws.String(tableName),
		"SUSPENDED":          aws.String("false"),
	}

	handlerProps := awslambdago.GoFunctionProps{
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String("lambda/subsuspendable/"),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(28)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Tracing:       awslambda.Tracing_ACTIVE,
		Environment:   &lambdaEnv,
	}

	handler := awslambdago.NewGoFunction(stack, aws.String(id), &handlerProps)

	eventSourceProps := awslambdaeventsources.SqsEventSourceProps{}
	handler.AddEventSource(awslambdaeventsources.NewSqsEventSource(queue, &eventSourceProps))

	return handler
}

func NewAPIGateway(stack awscdk.Stack, handler awslambdago.GoFunction) awsapigateway.LambdaRestApi {
	stageOptions := awsapigateway.StageOptions{
		StageName:        aws.String("prod"),
		LoggingLevel:     awsapigateway.MethodLoggingLevel_ERROR,
		TracingEnabled:   aws.Bool(true),
		MetricsEnabled:   aws.Bool(true),
		DataTraceEnabled: aws.Bool(true),
	}

	restApiProps := awsapigateway.LambdaRestApiProps{
		Handler:       handler,
		DeployOptions: &stageOptions,
	}

	return awsapigateway.NewLambdaRestApi(stack, aws.String(pubEndpointId), &restApiProps)
}

func NewSQSWorkshopStack(scope constructs.Construct, id string, props *CdkWorkshopStackProps) (stack awscdk.Stack) {
	var stackProps awscdk.StackProps

	//	stack...
	if props != nil {
		stackProps = props.StackProps
	}

	stack = awscdk.NewStack(scope, &id, &stackProps)

	// queue...
	queue1 := NewQueue(stack, queue1Name) // continuous sub
	queue2 := NewQueue(stack, queue2Name) // suspendable sub
	queue3 := NewQueue(stack, queue3Name) // no sub

	// topic...
	topic := NewTopic(stack, topicId, topicName)

	subProps := awssnssubscriptions.SqsSubscriptionProps{
		RawMessageDelivery: aws.Bool(true),
	}
	topic.AddSubscription(awssnssubscriptions.NewSqsSubscription(queue1, &subProps))
	topic.AddSubscription(awssnssubscriptions.NewSqsSubscription(queue2, &subProps))
	topic.AddSubscription(awssnssubscriptions.NewSqsSubscription(queue3, &subProps))

	// pub lambda...
	pubHandler := NewPubHandler(stack, pubHandlerId, topic)
	topic.GrantPublish(pubHandler)

	// sub lambdas...
	continuousSubHandler := NewContinuousSubHandler(stack, continuousSubHandlerId, queue1)
	queue1.GrantConsumeMessages(continuousSubHandler)

	suspendableSubHandler := NewSuspendableSubHandler(stack, suspendableSubHandlerId, queue2)
	queue2.GrantConsumeMessages(suspendableSubHandler)

	// gateway...
	NewAPIGateway(stack, pubHandler)

	// table...
	table := NewMessageTable(stack, tableId, tableName)
	table.GrantReadWriteData(continuousSubHandler)
	table.GrantReadWriteData(suspendableSubHandler)

	return stack
}

func NewCloudwatchDashboard(stack awscdk.Stack) awscloudwatch.Dashboard {
	return awscloudwatch.NewDashboard(stack, aws.String("eventsDashboard"), &awscloudwatch.DashboardProps{
		DashboardName:   aws.String("SQS1-" + *stack.Region()),
		DefaultInterval: awscdk.Duration_Hours(aws.Float64(24)),
	})
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)
	NewSQSWorkshopStack(app, stackId, &CdkWorkshopStackProps{})

	app.Synth(nil)
}
