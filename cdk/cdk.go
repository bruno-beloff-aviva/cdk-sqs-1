package main

// https://github.com/aviva-verde/cdk-standards.git
// https://docs.aws.amazon.com/cdk/v2/guide/resources.html
// https://docs.aws.amazon.com/code-library/latest/ug/go_2_sqs_code_examples.html
// https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_lambda_event_sources.SqsEventSource.html

import (
	"sqstest/cdk/gatewayhandler"
	"sqstest/cdk/snshandler"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// TODO: build a dashboard

const project = "SQS1"
const version = "0.2.0"

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

type CdkWorkshopStackProps struct {
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

func setupPubHandler(stack awscdk.Stack, props gatewayhandler.GatewayHandlerProps, topic awssns.Topic) {
	handler := gatewayhandler.GatewayHandler{
		EndpointId:        pubEndpointId,
		HandlerId:         pubHandlerId,
		SubscriptionTopic: topic,
		Entry:             "lambda/pub/",
		Environment: map[string]*string{
			"VERSION":   aws.String(version),
			"TOPIC_ARN": topic.TopicArn(),
		},
	}

	handler.Setup(stack, props)
}

func setupContinuousSubHandler(stack awscdk.Stack, props snshandler.SNSHandlerProps, topic awssns.Topic) {
	handler := snshandler.SNSHandler{
		SubscriptionTopic: topic,
		QueueName:         queue1Name,
		HandlerId:         continuousSubHandlerId,
		Entry:             "lambda/subcontinuous/",
		Environment: map[string]*string{
			"VERSION":            aws.String(version),
			"MESSAGE_TABLE_NAME": aws.String(tableName),
		},
	}

	handler.Setup(stack, props)
}

func setupSuspendableSubHandler(stack awscdk.Stack, props snshandler.SNSHandlerProps, topic awssns.Topic) {
	handler := snshandler.SNSHandler{
		SubscriptionTopic: topic,
		QueueName:         queue2Name,
		HandlerId:         suspendableSubHandlerId,
		Entry:             "lambda/subsuspendable/",
		Environment: map[string]*string{
			"VERSION":            aws.String(version),
			"MESSAGE_TABLE_NAME": aws.String(tableName),
			"SUSPENDED":          aws.String("false"),
		},
	}

	handler.Setup(stack, props)
}

func setupEmptySubHandler(stack awscdk.Stack, props snshandler.SNSHandlerProps, topic awssns.Topic) {
	handler := snshandler.SNSHandler{
		SubscriptionTopic: topic,
		QueueName:         queue3Name,
	}

	handler.Setup(stack, props)
}

func NewSQSWorkshopStack(scope constructs.Construct, id string, props *CdkWorkshopStackProps) (stack awscdk.Stack) {
	var stackProps awscdk.StackProps

	//	stack...
	if props != nil {
		stackProps = props.StackProps
	}

	stack = awscdk.NewStack(scope, &id, &stackProps)

	// topic...
	topic := NewTopic(stack, topicId, topicName)

	// table...
	table := NewMessageTable(stack, tableId, tableName)

	// key...
	keyProps := awskms.KeyProps{
		Alias:             aws.String("QueueKey"),
		EnableKeyRotation: aws.Bool(true),
	}

	queueKey := awskms.NewKey(stack, aws.String("Key"), &keyProps)

	// dashboard...
	dashboard := NewCloudwatchDashboard(stack)

	// pub lambda...
	pubProps := gatewayhandler.GatewayHandlerProps{
		CloudwatchDashboard: dashboard,
	}

	setupPubHandler(stack, pubProps, topic)

	// sub lambdas...
	subProps := snshandler.SNSHandlerProps{
		QueueKey:            queueKey,
		QueueMaxRetries:     queueMaxRetries,
		MessageTable:        table,
		CloudwatchDashboard: dashboard,
	}

	setupContinuousSubHandler(stack, subProps, topic)
	setupSuspendableSubHandler(stack, subProps, topic)
	setupEmptySubHandler(stack, subProps, topic)

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
