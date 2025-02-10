// cdk deploy --profile bb

// https://github.com/aviva-verde/cdk-standards.git
// https://docs.aws.amazon.com/cdk/v2/guide/resources.html

package main

import (
	sqs "sqstest/sqsaviva"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	awslambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const project = "SQS1"
const version = "0.0.2"

// const region = "eu-west-2"

const queueName = "TestQueue"

// const queueId = project + queueName

const handlerId = project + "PublishHandler"
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

// func NewHelloBucket(stack awscdk.Stack, id string, name string) awss3.IBucket {
// 	logConfig := s3.BucketLogConfiguration{
// 		BucketName: name,
// 		Region:     region,
// 		LogPrefix:  logPrefix,
// 	}

// 	props := s3.BucketProps{
// 		Stack:              stack,
// 		Name:               id,
// 		OverrideBucketName: aws.String(name),
// 		Versioned:          false,
// 		EventBridgeEnabled: false,
// 		LogConfiguration:   logConfig,
// 	}

// 	return s3.NewPrivateS3Bucket(props)
// }

func NewTestQueue(stack awscdk.Stack) awssqs.IQueue {
	queueKey := awskms.NewKey(stack, aws.String("queueKey"), &awskms.KeyProps{
		Alias:             jsii.String("testQueueKey"),
		EnableKeyRotation: jsii.Bool(true),
	})

	documentDeliveryQueue := sqs.NewSqsQueueWithDLQ(sqs.SqsQueueWithDLQProps{
		Stack:                    stack,
		QueueName:                "TestQueue",
		SQSKey:                   queueKey,
		QMaxReceiveCount:         3,
		QAlarmPeriod:             1,
		QAlarmThreshold:          1,
		QAlarmEvaluationPeriod:   1,
		DLQAlarmPeriod:           1,
		DLQAlarmThreshold:        1,
		DLQAlarmEvaluationPeriod: 1,
	})

	return documentDeliveryQueue
}

func NewPublishHandler(stack awscdk.Stack, lambdaEnv map[string]*string) awslambdago.GoFunction {
	publishHandler := awslambdago.NewGoFunction(stack, aws.String(handlerId), &awslambdago.GoFunctionProps{
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String("lambda/"),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(29)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Environment:   &lambdaEnv,
	})

	// Environment: &map[string]*string{
	// 	"DOCUMENT_DELIVERY_QUEUE_URL": documentDeliveryQueue.QueueUrl(),
	// },

	return publishHandler
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

	// lambda...
	lambdaEnv := map[string]*string{
		"VERSION":   aws.String(version),
		"QUEUE_URL": queue.QueueUrl(),
	}

	publishHandler := NewPublishHandler(stack, lambdaEnv)

	queue.GrantSendMessages(publishHandler)

	// bucket...
	// bucket := NewHelloBucket(stack, bucketId, bucketName)
	// bucket.GrantRead(helloHandler, nil)

	// table...
	// table := NewHitsTable(stack, tableId, tableName)
	// table.GrantReadWriteData(helloHandler)

	// gateway...
	restApiProps := awsapigateway.LambdaRestApiProps{Handler: publishHandler}
	awsapigateway.NewLambdaRestApi(stack, aws.String(endpointId), &restApiProps)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)
	NewSQSWorkshopStack(app, stackId, &CdkWorkshopStackProps{})

	app.Synth(nil)
}
