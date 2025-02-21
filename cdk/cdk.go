package main

// https://github.com/aviva-verde/cdk-standards.git
// https://docs.aws.amazon.com/cdk/v2/guide/resources.html
// https://docs.aws.amazon.com/code-library/latest/ug/go_2_sqs_code_examples.html
// https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_lambda_event_sources.SqsEventSource.html

import (
	"fmt"
	"sqstest/cdk/dashboard"
	"sqstest/cdk/gatewayhandler"
	"sqstest/cdk/snshandler"
	"sqstest/cdk/stackprops"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const account = "673007244143"
const region = "eu-west-2"

const project = "SQS1"
const version = "0.2.12"

var queueKeyId = project + "QueueKey"
var queueKeyAlias = "QueueKeyLive"

const queue1Name = "TestQueue1"
const queue2Name = "TestQueue2"
const queue3Name = "TestQueue3"
const queueMaxRetries = 3

const tableName = "TestMessageTableV2"
const tableId = project + tableName

const topicName = "TestTopic"
const topicId = project + topicName

const pubHandlerId = project + "PubHandler"
const pubEndpointId = project + "PubEndpoint"

const continuousSubHandlerId = project + "ContinuousHandler"
const suspendableSubHandlerId = project + "SudspendableHandler"

const stackId = project + "Stack"

const dashboardId = project + "Dashboard"

// TODO: don't create a queue key if it already exists.

func NewSQSStack(scope constructs.Construct, id string, stackProps *stackprops.CdkStackProps) (stack awscdk.Stack) {
	stack = stackProps.NewStack(scope, id)

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

	c0 := setupPubHandler(stack, *stackProps, pubProps, topic)

	// sub lambdas...
	subProps := snshandler.SNSCommonProps{
		QueueKey:        queueKey,
		QueueMaxRetries: queueMaxRetries,
		MessageTable:    table,
		Dashboard:       dash,
	}

	c1 := setupContinuousSubHandler(stack, subProps, topic)
	c2 := setupSuspendableSubHandler(stack, subProps, topic)
	c3 := setupEmptySubHandler(stack, subProps, topic)

	// dashboard widgets...
	dash.AddWidgetsRow(c0.GatewayMetricsGraphWidget(), c0.LambdaMetricsGraphWidget(), c1.LambdaMetricsGraphWidget(), c2.LambdaMetricsGraphWidget())
	dash.AddWidgetsRow(c1.QueueMetricsGraphWidget(), c1.DLQMetricsGraphWidget(), c2.QueueMetricsGraphWidget(), c2.DLQMetricsGraphWidget())
	dash.AddWidgetsRow(c0.TopicMetricsGraphWidget(), c3.QueueMetricsGraphWidget(), c3.DLQMetricsGraphWidget())

	return stack
}

func setupDashboard(stack awscdk.Stack) dashboard.Dashboard {
	dash := dashboard.NewDashboard(stack, dashboardId)

	return dash
}

func setupMessageTable(stack awscdk.Stack, id string, name string) awsdynamodb.ITable {
	tableProps := awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{Name: aws.String("PK"), Type: awsdynamodb.AttributeType_STRING},
		SortKey:      &awsdynamodb.Attribute{Name: aws.String("Received"), Type: awsdynamodb.AttributeType_STRING},
		TableName:    aws.String(name),
	}

	return awsdynamodb.NewTable(stack, aws.String(id), &tableProps)
}

func setupTopic(stack awscdk.Stack, id string, name string) gatewayhandler.NamedTopic {
	topicProps := awssns.TopicProps{
		DisplayName: aws.String(name),
	}

	topic := gatewayhandler.NamedTopic{
		Topic: awssns.NewTopic(stack, aws.String(id), &topicProps),
		Name:  name,
	}

	return topic
}

func setupQueueKey(stack awscdk.Stack) awskms.IKey {
	keyProps := awskms.KeyProps{
		Alias:             aws.String(queueKeyAlias),
		EnableKeyRotation: aws.Bool(true),
	}

	// *** queue alias ARN: arn:${Token[AWS.Partition.5]}:kms:${Token[AWS.Region.6]}:${Token[AWS.AccountId.2]}:SQS1QueueKey

	// alias := awskms.Alias_FromAliasName(stack, aws.String(queueKeyAlias), aws.String(queueKeyId))
	// fmt.Printf("*** queue alias ARN: %v\n", *alias.KeyArn())

	key := awskms.Key_FromLookup(stack, &queueKeyId, &awskms.KeyLookupOptions{
		AliasName:               aws.String(queueKeyAlias),
		ReturnDummyKeyOnMissing: aws.Bool(false),
	})

	fmt.Printf("*** found key: %v\n", key)

	return awskms.NewKey(stack, aws.String(queueKeyId), &keyProps)
}

func setupPubHandler(stack awscdk.Stack, stackProps stackprops.CdkStackProps, commonProps gatewayhandler.GatewayCommonProps, topic gatewayhandler.NamedTopic) gatewayhandler.GatewayConstruct {
	environment := map[string]*string{
		"VERSION":   aws.String(version),
		"TOPIC_ARN": topic.TopicArn(),
	}

	builder := gatewayhandler.GatewayBuilder{
		EndpointId:       pubEndpointId,
		HandlerId:        pubHandlerId,
		PublicationTopic: topic,
		Entry:            "lambda/pub/",
		Environment:      environment,
	}

	return builder.Setup(stack, stackProps, commonProps)
}

func setupContinuousSubHandler(stack awscdk.Stack, commonProps snshandler.SNSCommonProps, topic awssns.Topic) snshandler.SNSConstruct {
	environment := map[string]*string{
		"VERSION":            aws.String(version),
		"MESSAGE_TABLE_NAME": aws.String(tableName),
	}

	builder := snshandler.SNSBuilder{
		SubscriptionTopic: topic,
		QueueName:         queue1Name,
		HandlerId:         continuousSubHandlerId,
		Entry:             "lambda/subcontinuous/",
		Environment:       environment,
	}

	return builder.Setup(stack, commonProps)
}

func setupSuspendableSubHandler(stack awscdk.Stack, commonProps snshandler.SNSCommonProps, topic awssns.Topic) snshandler.SNSConstruct {
	environment := map[string]*string{
		"VERSION":            aws.String(version),
		"MESSAGE_TABLE_NAME": aws.String(tableName),
		"SUSPENDED":          aws.String("false"),
	}

	builder := snshandler.SNSBuilder{
		SubscriptionTopic: topic,
		QueueName:         queue2Name,
		HandlerId:         suspendableSubHandlerId,
		Entry:             "lambda/subsuspendable/",
		Environment:       environment,
	}

	return builder.Setup(stack, commonProps)
}

func setupEmptySubHandler(stack awscdk.Stack, commonProps snshandler.SNSCommonProps, topic awssns.Topic) snshandler.SNSConstruct {
	builder := snshandler.SNSBuilder{
		SubscriptionTopic: topic,
		QueueName:         queue3Name,
	}

	return builder.Setup(stack, commonProps)
}

// panic: Error: Cannot retrieve value from context provider key-provider since account/region are not specified at the stack level.
// Configure "env" with an account and region when you define your stack.
// See https://docs.aws.amazon.com/cdk/latest/guide/environments.html for more details.

func main() {
	defer jsii.Close()

	env := awscdk.Environment{
		Account: aws.String(account),
		Region:  aws.String(region),
	}

	stackProps := awscdk.StackProps{
		Env: &env,
	}

	cdkStackProps := &stackprops.CdkStackProps{
		StackProps: stackProps,
		Version:    version,
	}

	app := awscdk.NewApp(nil)
	NewSQSStack(app, stackId, cdkStackProps)

	app.Synth(nil)
}
