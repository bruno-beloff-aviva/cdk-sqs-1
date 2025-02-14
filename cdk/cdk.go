package main

// https://github.com/aviva-verde/cdk-standards.git
// https://docs.aws.amazon.com/cdk/v2/guide/resources.html
// https://docs.aws.amazon.com/code-library/latest/ug/go_2_sqs_code_examples.html
// https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_lambda_event_sources.SqsEventSource.html

import (
	sqs "sqstest/aviva/sqs"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
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

const project = "SQS1"
const version = "0.1.4"

const queueName = "TestQueue"
const queueMaxRetries = 3

const tableName = "TestMessageTableV1"
const tableId = project + tableName

const topicName = project + "TestTopic"
const topicId = project + topicName

const pubHandlerId = project + "PubHandler"
const pubEndpointId = project + "PubEndpoint"

const continuousSubHandlerId = project + "ContinuousHandler"
const suspendableSubHandlerId = project + "SudspendableHandler"

const stackId = project + "Stack"

type CdkWorkshopStackProps struct {
	awscdk.StackProps
}

func NewMessageTable(scope constructs.Construct, id string, name string) awsdynamodb.ITable {
	tableProps := awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{Name: aws.String("Sent"), Type: awsdynamodb.AttributeType_STRING},
		SortKey:      &awsdynamodb.Attribute{Name: aws.String("Path"), Type: awsdynamodb.AttributeType_STRING},
		TableName:    aws.String(name),
	}

	return awsdynamodb.NewTable(scope, aws.String(id), &tableProps)
}

func NewTopic(stack awscdk.Stack) awssns.Topic {
	topicProps := awssns.TopicProps{
		DisplayName: aws.String(topicName),
	}

	return awssns.NewTopic(stack, aws.String(topicId), &topicProps)
}

// TODO: each lambda must have its own queue
func NewTestQueue(stack awscdk.Stack) awssqs.IQueue {
	queueKey := awskms.NewKey(stack, aws.String("queueKey"), &awskms.KeyProps{
		Alias:             jsii.String("testQueueKey"),
		EnableKeyRotation: jsii.Bool(true),
	})

	queueProps := sqs.SqsQueueWithDLQProps{
		Stack:                    stack,
		QueueName:                queueName,
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

func NewPubHandler(stack awscdk.Stack, topic awssns.Topic) awslambdago.GoFunction {
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

	return awslambdago.NewGoFunction(stack, aws.String(pubHandlerId), &handlerProps)
}

func NewContinuousSubHandler(stack awscdk.Stack) awslambdago.GoFunction {
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

	return awslambdago.NewGoFunction(stack, aws.String(continuousSubHandlerId), &handlerProps)
}

func NewSuspendableSubHandler(stack awscdk.Stack) awslambdago.GoFunction {
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

	return awslambdago.NewGoFunction(stack, aws.String(suspendableSubHandlerId), &handlerProps)
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
	queue := NewTestQueue(stack) //	we need two queues

	// topic...
	topic := NewTopic(stack)

	subProps := awssnssubscriptions.SqsSubscriptionProps{}
	topic.AddSubscription(awssnssubscriptions.NewSqsSubscription(queue, &subProps))

	// lambdas...
	pubHandler := NewPubHandler(stack, topic)
	topic.GrantPublish(pubHandler)

	eventSourceProps := awslambdaeventsources.SqsEventSourceProps{}

	continuousSubHandler := NewContinuousSubHandler(stack)
	continuousSubHandler.AddEventSource(awslambdaeventsources.NewSqsEventSource(queue, &eventSourceProps))
	queue.GrantConsumeMessages(continuousSubHandler)

	suspendableSubHandler := NewSuspendableSubHandler(stack)
	suspendableSubHandler.AddEventSource(awslambdaeventsources.NewSqsEventSource(queue, &eventSourceProps))
	queue.GrantConsumeMessages(suspendableSubHandler)

	// gateway...
	NewAPIGateway(stack, pubHandler)

	// topic...
	topic.GrantPublish(pubHandler)

	// table...
	table := NewMessageTable(stack, tableId, tableName)
	table.GrantReadWriteData(continuousSubHandler)
	table.GrantReadWriteData(suspendableSubHandler)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)
	NewSQSWorkshopStack(app, stackId, &CdkWorkshopStackProps{})

	app.Synth(nil)
}
