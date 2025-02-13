// https://github.com/aviva-verde/cdk-standards.git
// https://docs.aws.amazon.com/cdk/v2/guide/resources.html
// https://docs.aws.amazon.com/code-library/latest/ug/go_2_sqs_code_examples.html
// https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_lambda_event_sources.SqsEventSource.html

package main

import (
	sqs "sqstest/aviva/sqs"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	awslambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const project = "SQS1"
const version = "0.1.2"

const queueName = "TestQueue"
const queueMaxRetries = 5

const tableName = "TestMessageTableV1"
const tableId = project + tableName

const publishHandlerId = project + "PublishHandler"
const publishEndpointId = project + "PublishEndpoint"

const subscribeHandlerId = project + "SubscribeHandler"

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

func NewTestQueue(stack awscdk.Stack) awssqs.IQueue {
	queueKey := awskms.NewKey(stack, aws.String("queueKey"), &awskms.KeyProps{
		Alias:             jsii.String("testQueueKey"),
		EnableKeyRotation: jsii.Bool(true),
	})

	queueProps := sqs.SqsQueueWithDLQProps{
		Stack:                    stack,
		QueueName:                queueName,
		SQSKey:                   queueKey,
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

func NewPublishHandler(stack awscdk.Stack, queue awssqs.IQueue) awslambdago.GoFunction {
	lambdaEnv := map[string]*string{
		"VERSION":   aws.String(version),
		"QUEUE_URL": queue.QueueUrl(),
	}

	handlerProps := awslambdago.GoFunctionProps{
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String("lambda/publish/"),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(28)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Tracing:       awslambda.Tracing_ACTIVE,
		Environment:   &lambdaEnv,
	}

	return awslambdago.NewGoFunction(stack, aws.String(publishHandlerId), &handlerProps)
}

func NewSubscribeHandler(stack awscdk.Stack) awslambdago.GoFunction {
	lambdaEnv := map[string]*string{
		"VERSION":            aws.String(version),
		"MESSAGE_TABLE_NAME": aws.String(tableName),
		"SUSPENDED":          aws.String("false"),
	}

	handlerProps := awslambdago.GoFunctionProps{
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String("lambda/suspendablesubscribe/"),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(28)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Tracing:       awslambda.Tracing_ACTIVE,
		Environment:   &lambdaEnv,
	}

	return awslambdago.NewGoFunction(stack, aws.String(subscribeHandlerId), &handlerProps)
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

	return awsapigateway.NewLambdaRestApi(stack, aws.String(publishEndpointId), &restApiProps)
}

func NewSQSWorkshopStack(scope constructs.Construct, id string, props *CdkWorkshopStackProps) (stack awscdk.Stack) {
	var stackProps awscdk.StackProps

	//	stack...
	if props != nil {
		stackProps = props.StackProps
	}

	stack = awscdk.NewStack(scope, &id, &stackProps)

	// queue...
	queue := NewTestQueue(stack)

	// lambdas...
	publishHandler := NewPublishHandler(stack, queue)
	queue.GrantSendMessages(publishHandler)

	subscribeHandler := NewSubscribeHandler(stack)
	eventSourceProps := awslambdaeventsources.SqsEventSourceProps{}
	subscribeHandler.AddEventSource(awslambdaeventsources.NewSqsEventSource(queue, &eventSourceProps))
	queue.GrantConsumeMessages(subscribeHandler)

	// gateway...
	NewAPIGateway(stack, publishHandler)

	// table...
	table := NewMessageTable(stack, tableId, tableName)
	table.GrantReadWriteData(subscribeHandler)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)
	NewSQSWorkshopStack(app, stackId, &CdkWorkshopStackProps{})

	app.Synth(nil)
}
