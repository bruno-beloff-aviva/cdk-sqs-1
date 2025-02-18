package gatewayhandler

import (
	"fmt"
	"sqstest/cdk/dashboard"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	awslambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/aws-sdk-go/aws"
)

const stage = "prod"

// common to all Gateway handlers
type GatewayCommonProps struct {
	Dashboard dashboard.Dashboard
}

// specific to an Gateway handler
type GatewayBuilder struct {
	EndpointId       string
	HandlerId        string
	PublicationTopic awssns.Topic
	Entry            string
	Environment      map[string]*string
}

type GatewayConstruct struct {
	Builder   GatewayBuilder
	Gateway   awsapigateway.LambdaRestApi
	Handler   awslambdago.GoFunction
	Dashboard dashboard.Dashboard
}

func (h GatewayBuilder) Setup(stack awscdk.Stack, props GatewayCommonProps) GatewayConstruct {
	var c GatewayConstruct

	c.Builder = h
	c.Dashboard = props.Dashboard
	c.Handler = h.setupPubHandler(stack)

	h.PublicationTopic.GrantPublish(c.Handler)
	h.setupGateway(stack, c.Handler)

	return c
}

func (h GatewayBuilder) setupPubHandler(stack awscdk.Stack) awslambdago.GoFunction {
	handlerProps := awslambdago.GoFunctionProps{
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String(h.Entry),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(28)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Tracing:       awslambda.Tracing_ACTIVE,
		Environment:   &h.Environment,
	}

	handler := awslambdago.NewGoFunction(stack, aws.String(h.HandlerId), &handlerProps)

	return handler
}

func (h GatewayBuilder) setupGateway(stack awscdk.Stack, handler awslambdago.GoFunction) awsapigateway.LambdaRestApi {
	stageOptions := awsapigateway.StageOptions{
		StageName:        aws.String(stage),
		LoggingLevel:     awsapigateway.MethodLoggingLevel_ERROR,
		TracingEnabled:   aws.Bool(true),
		MetricsEnabled:   aws.Bool(true),
		DataTraceEnabled: aws.Bool(true),
	}

	restApiProps := awsapigateway.LambdaRestApiProps{
		Handler:       handler,
		DeployOptions: &stageOptions,
	}

	return awsapigateway.NewLambdaRestApi(stack, aws.String(h.EndpointId), &restApiProps)
}

func (c GatewayConstruct) LambdaMetricsGraphWidget() awscloudwatch.GraphWidget {
	region := c.Handler.Stack().Region()

	invocationsMetric := c.Dashboard.CreateLambdaMetric(*region, "Invocations", c.Handler.FunctionName(), "Sum")
	errorsMetric := c.Dashboard.CreateLambdaMetric(*region, "Errors", c.Handler.FunctionName(), "Sum")
	metrics := []awscloudwatch.IMetric{invocationsMetric, errorsMetric}

	return c.Dashboard.CreateGraphWidget(*region, fmt.Sprintf("%s - Invocations and Errors", c.Builder.HandlerId), metrics)
}

func (c GatewayConstruct) GatewayMetricsGraphWidget() awscloudwatch.GraphWidget {
	region := c.Handler.Stack().Region()

	invocationsMetric := c.Dashboard.CreateGatewayMetric(*region, "Count", c.Builder.EndpointId, stage, "Sum")
	errorsMetric := c.Dashboard.CreateGatewayMetric(*region, "5XXError", c.Builder.EndpointId, stage, "Sum")
	metrics := []awscloudwatch.IMetric{invocationsMetric, errorsMetric}

	return c.Dashboard.CreateGraphWidget(*region, fmt.Sprintf("%s - Invocations and Errors", c.Builder.EndpointId), metrics)
}

func (c GatewayConstruct) TopicMetricsGraphWidget() awscloudwatch.GraphWidget {
	region := c.Handler.Stack().Region()

	publicationsMetric := c.Dashboard.CreateTopicMetric(*region, "NumberOfMessagesPublished", c.Builder.PublicationTopic.TopicName(), "Sum")
	failsMetric := c.Dashboard.CreateTopicMetric(*region, "NumberOfNotificationsFailed", c.Builder.PublicationTopic.TopicName(), "Sum")
	metrics := []awscloudwatch.IMetric{publicationsMetric, failsMetric}

	return c.Dashboard.CreateGraphWidget(*region, fmt.Sprintf("%s - Publications and Failures", c.Builder.PublicationTopic), metrics)
}
