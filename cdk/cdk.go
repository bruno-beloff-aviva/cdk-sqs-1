package main

// https://github.com/aviva-verde/cdk-standards.git
// https://docs.aws.amazon.com/cdk/v2/guide/resources.html
// https://docs.aws.amazon.com/code-library/latest/ug/go_2_sqs_code_examples.html
// https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_lambda_event_sources.SqsEventSource.html

import (
	"sqstest/cdk/dashboard"
	"sqstest/cdk/gatewayhandler"
	"sqstest/cdk/snshandler"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// TODO: AddAlias for lambdas

const project = "SQS1"
const version = "0.2.1"

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

func setupDashboard(stack awscdk.Stack) dashboard.Dashboard {
	dash := dashboard.NewDashboard(stack, "SQS1")

	return dash
}

func setupMessageTable(stack awscdk.Stack, id string, name string) awsdynamodb.ITable {
	tableProps := awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{Name: aws.String("PK"), Type: awsdynamodb.AttributeType_STRING},
		SortKey:      &awsdynamodb.Attribute{Name: aws.String("Path"), Type: awsdynamodb.AttributeType_STRING},
		TableName:    aws.String(name),
	}

	return awsdynamodb.NewTable(stack, aws.String(id), &tableProps)
}

func setupTopic(stack awscdk.Stack, id string, name string) awssns.Topic {
	topicProps := awssns.TopicProps{
		DisplayName: aws.String(name),
	}

	return awssns.NewTopic(stack, aws.String(id), &topicProps)
}

func setupQueueKey(stack awscdk.Stack) awskms.IKey {
	keyProps := awskms.KeyProps{
		Alias:             aws.String("QueueKey"),
		EnableKeyRotation: aws.Bool(true),
	}

	return awskms.NewKey(stack, aws.String("Key"), &keyProps)
}

func setupPubHandler(stack awscdk.Stack, props gatewayhandler.GatewayCommonProps, topic awssns.Topic) gatewayhandler.GatewayConstruct {
	builder := gatewayhandler.GatewayBuilder{
		EndpointId:        pubEndpointId,
		HandlerId:         pubHandlerId,
		SubscriptionTopic: topic,
		Entry:             "lambda/pub/",
		Environment: map[string]*string{
			"VERSION":   aws.String(version),
			"TOPIC_ARN": topic.TopicArn(),
		},
	}

	return builder.Setup(stack, props)
}

func setupContinuousSubHandler(stack awscdk.Stack, props snshandler.SNSCommonProps, topic awssns.Topic) snshandler.SNSConstruct {
	builder := snshandler.SNSBuilder{
		SubscriptionTopic: topic,
		QueueName:         queue1Name,
		HandlerId:         continuousSubHandlerId,
		Entry:             "lambda/subcontinuous/",
		Environment: map[string]*string{
			"VERSION":            aws.String(version),
			"MESSAGE_TABLE_NAME": aws.String(tableName),
		},
	}

	return builder.Setup(stack, props)
}

func setupSuspendableSubHandler(stack awscdk.Stack, props snshandler.SNSCommonProps, topic awssns.Topic) snshandler.SNSConstruct {
	builder := snshandler.SNSBuilder{
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

	return builder.Setup(stack, props)
}

func setupEmptySubHandler(stack awscdk.Stack, props snshandler.SNSCommonProps, topic awssns.Topic) snshandler.SNSConstruct {
	builder := snshandler.SNSBuilder{
		SubscriptionTopic: topic,
		QueueName:         queue3Name,
	}

	return builder.Setup(stack, props)
}

func NewSQSWorkshopStack(scope constructs.Construct, id string, props *CdkWorkshopStackProps) (stack awscdk.Stack) {
	var stackProps awscdk.StackProps

	//	stack...
	if props != nil {
		stackProps = props.StackProps
	}

	stack = awscdk.NewStack(scope, &id, &stackProps)

	// dashboard...
	dash := setupDashboard(stack)

	// topic...
	topic := setupTopic(stack, topicId, topicName)

	// table...
	table := setupMessageTable(stack, tableId, tableName)

	// key...
	queueKey := setupQueueKey(stack)

	// pub lambda...
	pubProps := gatewayhandler.GatewayCommonProps{
		Dashboard: dash,
	}

	setupPubHandler(stack, pubProps, topic)

	// sub lambdas...
	subProps := snshandler.SNSCommonProps{
		QueueKey:        queueKey,
		QueueMaxRetries: queueMaxRetries,
		MessageTable:    table,
		Dashboard:       dash,
	}

	c := setupContinuousSubHandler(stack, subProps, topic)
	setupSuspendableSubHandler(stack, subProps, topic)
	setupEmptySubHandler(stack, subProps, topic)

	dash.AddLambdaMetrics(*stack.Region(), c.Build.HandlerId) // TODO: put on construct

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)
	NewSQSWorkshopStack(app, stackId, &CdkWorkshopStackProps{})

	app.Synth(nil)
}
