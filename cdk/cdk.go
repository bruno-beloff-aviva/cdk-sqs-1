package main

// https://github.com/aviva-verde/cdk-standards.git
// https://docs.aws.amazon.com/cdk/v2/guide/resources.html
// https://docs.aws.amazon.com/code-library/latest/ug/go_2_sqs_code_examples.html
// https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_lambda_event_sources.SqsEventSource.html

import (
	"sqstest/cdk/dashboard"
	"sqstest/cdk/eventhandler"
	"sqstest/cdk/gatewayhandler"
	"sqstest/cdk/stackprops"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/constructs-go/constructs/v10"
)

const (
	project                 = "SQS1"
	version                 = "0.2.25"
	queueKeyAlias           = "QueueKeyLive"
	queue1Name              = "TestQueue1"
	queue2Name              = "TestQueue2"
	queue3Name              = "TestQueue3"
	queueMaxRetries         = 3
	eventBusName            = "TestEventBus"
	tableName               = "TestMessageTableV2"
	topicName               = "TestTopic"
	eventBusId              = project + eventBusName
	tableId                 = project + tableName
	queueKeyId              = project + "QueueKey"
	topicId                 = project + topicName
	pubHandlerId            = project + "PubHandler"
	pubEndpointId           = project + "PubEndpoint"
	continuousSubHandlerId  = project + "ContinuousHandler"
	suspendableSubHandlerId = project + "SudspendableHandler"
	stackId                 = project + "Stack"
	dashboardId             = project + "Dashboard"
)

func NewSQSStack(scope constructs.Construct, id string, stackProps *stackprops.CdkStackProps) (stack awscdk.Stack) {
	stack = stackProps.NewStack(scope, id)

	// dashboard...
	dash := setupDashboard(stack)

	// event bus...
	eventBus := setupEventBus(stack)

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

	c0 := setupPubHandler(stack, *stackProps, pubProps, eventBus, topic)
	eventBus.GrantPutEventsTo(c0.Handler)

	// sub lambdas...
	subProps := eventhandler.EventHandlerCommonProps{
		QueueKey:        queueKey,
		QueueMaxRetries: queueMaxRetries,
		MessageTable:    table,
		Dashboard:       dash,
	}

	c1 := setupContinuousSubHandler(stack, subProps, topic)
	c2 := setupSuspendableSubHandler(stack, subProps, topic)
	c3 := setupEmptySubHandler(stack, subProps, topic)

	rule, targetInput := setupEventBusRule(stack, eventBus, pubEndpointId)

	rule.AddTarget(awseventstargets.NewSqsQueue(c1.Queue, &targetInput))
	rule.AddTarget(awseventstargets.NewSqsQueue(c2.Queue, &targetInput))
	rule.AddTarget(awseventstargets.NewSqsQueue(c3.Queue, &targetInput))

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

func setupTopic(stack awscdk.Stack, id string, name string) gatewayhandler.NamedTopic { // TODO: remove this?
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

	return awskms.NewKey(stack, aws.String(queueKeyId), &keyProps)
}

func setupEventBus(stack awscdk.Stack) awsevents.IEventBus {
	busProps := awsevents.EventBusProps{
		EventBusName: aws.String(eventBusName),
		DeadLetterQueue: awssqs.NewQueue(stack, aws.String(eventBusId+"DLQ"), &awssqs.QueueProps{
			QueueName: aws.String(eventBusName + "DLQ"),
		}),
	}

	bus := awsevents.NewEventBus(stack, aws.String(eventBusId), &busProps)

	return bus
}

// https://dev.to/chrisarmstrong/sqs-queues-as-an-eventbridge-rule-target-3d2g

func setupEventBusRule(stack awscdk.Stack, eventBus awsevents.IEventBus, source string) (rule awsevents.Rule, targetInput awseventstargets.SqsQueueProps) {
	eventPattern := &awsevents.EventPattern{
		Source: &[]*string{
			aws.String(source),
		},
	}

	ruleProps := awsevents.RuleProps{
		EventBus:     eventBus,
		EventPattern: eventPattern,
		RuleName:     aws.String("GatewayTestMessageRule"),
	}

	rule = awsevents.NewRule(stack, aws.String(project+"GatewayTestMessageRule"), &ruleProps) // TODO: sort out name for rule

	targetInput = awseventstargets.SqsQueueProps{
		Message:        awsevents.RuleTargetInput_FromEventPath(aws.String("$.detail")),
		MessageGroupId: aws.String("messageGroupId"),
	}

	return rule, targetInput
}

func setupPubHandler(stack awscdk.Stack, stackProps stackprops.CdkStackProps, commonProps gatewayhandler.GatewayCommonProps, eventBus awsevents.IEventBus, topic gatewayhandler.NamedTopic) gatewayhandler.GatewayConstruct {
	environment := map[string]*string{
		"VERSION":        aws.String(version),
		"EVENT_SOURCE":   aws.String(pubEndpointId),
		"EVENT_BUS_NAME": eventBus.EventBusName(),
	}

	builder := gatewayhandler.GatewayBuilder{
		EndpointId:       pubEndpointId,
		HandlerId:        pubHandlerId,
		EventBus:         eventBus,
		PublicationTopic: topic,
		Entry:            "lambda/pub/",
		Environment:      environment,
	}

	return builder.Setup(stack, stackProps, commonProps)
}

func setupContinuousSubHandler(stack awscdk.Stack, commonProps eventhandler.EventHandlerCommonProps, topic awssns.Topic) eventhandler.EventHandlerConstruct {
	environment := map[string]*string{
		"VERSION":            aws.String(version),
		"MESSAGE_TABLE_NAME": aws.String(tableName),
	}

	builder := eventhandler.EventHandlerBuilder{
		SubscriptionTopic: topic,
		QueueName:         queue1Name,
		HandlerId:         continuousSubHandlerId,
		Entry:             "lambda/subcontinuous/",
		Environment:       environment,
	}

	return builder.Setup(stack, commonProps)
}

func setupSuspendableSubHandler(stack awscdk.Stack, commonProps eventhandler.EventHandlerCommonProps, topic awssns.Topic) eventhandler.EventHandlerConstruct {
	environment := map[string]*string{
		"VERSION":            aws.String(version),
		"MESSAGE_TABLE_NAME": aws.String(tableName),
		"SUSPENDED":          aws.String("false"),
	}

	builder := eventhandler.EventHandlerBuilder{
		SubscriptionTopic: topic,
		QueueName:         queue2Name,
		HandlerId:         suspendableSubHandlerId,
		Entry:             "lambda/subsuspendable/",
		Environment:       environment,
	}

	return builder.Setup(stack, commonProps)
}

func setupEmptySubHandler(stack awscdk.Stack, commonProps eventhandler.EventHandlerCommonProps, topic awssns.Topic) eventhandler.EventHandlerConstruct {
	builder := eventhandler.EventHandlerBuilder{
		SubscriptionTopic: topic,
		QueueName:         queue3Name,
	}

	return builder.Setup(stack, commonProps)
}

func main() {
	cdkStackProps := stackprops.CdkStackProps{
		StackProps: awscdk.StackProps{},
		Version:    version,
	}

	app := awscdk.NewApp(nil)
	NewSQSStack(app, stackId, &cdkStackProps)

	app.Synth(nil)
}
