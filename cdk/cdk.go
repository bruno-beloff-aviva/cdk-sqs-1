// https://github.com/aviva-verde/cdk-standards.git
// https://docs.aws.amazon.com/cdk/v2/guide/resources.html
// https://docs.aws.amazon.com/code-library/latest/ug/go_2_sqs_code_examples.html
// https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_lambda_event_sources.SqsEventSource.html

package main

import (
	apigateway "sqstest/aviva/apigateway"
	sqs "sqstest/aviva/sqs"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
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
const version = "0.0.8"

const queueName = "TestQueue"
const publishHandlerId = project + "PublishHandler"
const subscribeHandlerId = project + "SubscribeHandler"
const endpointId = project + "PublishEndpoint"
const stackId = project + "Stack"

type CdkWorkshopStackProps struct {
	awscdk.StackProps
}

// func NewHitsTable(scope constructs.Construct, id string, name string) awsdynamodb.ITable {
// 	this := constructs.NewConstruct(scope, &id)

// 	// keep ID different from name at this stage, to prevent "already exists" panic
// 	table := awsdynamodb.NewTable(this, aws.String(id), &awsdynamodb.TableProps{
// 		PartitionKey: &awsdynamodb.Attribute{Name: aws.String("path"), Type: awsdynamodb.AttributeType_STRING},
// 		TableName:    aws.String(name),
// 	})

// 	return table
// }

func NewTestQueue(stack awscdk.Stack) awssqs.IQueue {
	queueKey := awskms.NewKey(stack, aws.String("queueKey"), &awskms.KeyProps{
		Alias:             jsii.String("testQueueKey"),
		EnableKeyRotation: jsii.Bool(true),
	})

	messageQueue := sqs.NewSqsQueueWithDLQ(sqs.SqsQueueWithDLQProps{
		Stack:                    stack,
		QueueName:                queueName,
		SQSKey:                   queueKey,
		QMaxReceiveCount:         10,
		QAlarmPeriod:             1,
		QAlarmThreshold:          1,
		QAlarmEvaluationPeriod:   1,
		DLQAlarmPeriod:           1,
		DLQAlarmThreshold:        1,
		DLQAlarmEvaluationPeriod: 1,
	})
	// Fifo:                     true,

	return messageQueue
}

func NewPublishHandler(stack awscdk.Stack, queue awssqs.IQueue) awslambdago.GoFunction {
	lambdaEnv := map[string]*string{
		"VERSION":   aws.String(version),
		"QUEUE_URL": queue.QueueUrl(),
	}

	handler := awslambdago.NewGoFunction(stack, aws.String(publishHandlerId), &awslambdago.GoFunctionProps{
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String("lambda/publish/"),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(28)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Environment:   &lambdaEnv,
	})

	return handler
}

func NewSubscribeHandler(stack awscdk.Stack) awslambdago.GoFunction {
	lambdaEnv := map[string]*string{
		"VERSION": aws.String(version),
	}

	handler := awslambdago.NewGoFunction(stack, aws.String(subscribeHandlerId), &awslambdago.GoFunctionProps{
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String("lambda/subscribe/"),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(28)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Environment:   &lambdaEnv,
	})

	return handler
}

func NewAPIGateway(stack awscdk.Stack, handler awslambdago.GoFunction) awsapigateway.LambdaRestApi {

	// https://stackoverflow.com/questions/61613801/cdk-deployment-api-gateway-cloudwatch-logs-role-arn-must-be-set-in-account-set

	corsOptions := awsapigateway.CorsOptions{
		AllowOrigins: awsapigateway.Cors_ALL_ORIGINS(),
		AllowMethods: awsapigateway.Cors_ALL_METHODS(),
	}

	endpointType := []awsapigateway.EndpointType{
		awsapigateway.EndpointType_PRIVATE,
	}

	props := apigateway.PublicAPIGatewayProps{}
	props.Stack = stack
	props.Name = endpointId
	props.Description = "API Gateway for SQS testing"
	props.DefaultHandler = handler

	props.DefaultCorsPreflightOptions = corsOptions
	props.EndpointTypes = &endpointType

	return apigateway.NewPublicAPIGateway(props)

	// restApiProps := awsapigateway.LambdaRestApiProps{Handler: handler}
	// return awsapigateway.NewLambdaRestApi(stack, aws.String(endpointId), &restApiProps)
}

func NewSQSWorkshopStack(scope constructs.Construct, id string, props *CdkWorkshopStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	//	stack...
	stack := awscdk.NewStack(scope, &id, &sprops)

	// queue...
	queue := NewTestQueue(stack)

	// lambdas...
	publishHandler := NewPublishHandler(stack, queue)
	queue.GrantSendMessages(publishHandler)

	subscribeHandler := NewSubscribeHandler(stack)
	subscribeHandler.AddEventSource(awslambdaeventsources.NewSqsEventSource(queue, &awslambdaeventsources.SqsEventSourceProps{}))
	queue.GrantConsumeMessages(subscribeHandler)

	// gateway...
	NewAPIGateway(stack, publishHandler)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)
	NewSQSWorkshopStack(app, stackId, &CdkWorkshopStackProps{})

	app.Synth(nil)
}
