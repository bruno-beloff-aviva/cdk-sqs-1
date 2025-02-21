package gatewayhandler

// https://www.youtube.com/watch?v=5v3rW2fPbLs
// https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_lambda.Alias.html
// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/lambda#Client.CreateAlias
// https://stackoverflow.com/questions/63477633/how-do-you-point-api-gateway-to-a-lambda-alias-in-cdk

// WARNING: switching from lambda to lambda alias required that the stack was destroyed and re-created

import (
	"fmt"
	"sqstest/cdk/dashboard"
	"sqstest/cdk/stackprops"

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

type NamedTopic struct {
	awssns.Topic
	Name string
}

// common to all Gateway handlers
type GatewayCommonProps struct {
	Version   string
	Dashboard dashboard.Dashboard
}

// specific to an Gateway handler
type GatewayBuilder struct {
	EndpointId       string
	HandlerId        string
	PublicationTopic NamedTopic
	Entry            string
	Environment      map[string]*string
}

type GatewayConstruct struct {
	Builder      GatewayBuilder
	Dashboard    dashboard.Dashboard
	HandlerAlias awslambda.Alias
	Gateway      awsapigateway.LambdaRestApi
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (b GatewayBuilder) Setup(stack awscdk.Stack, stackProps stackprops.CdkStackProps, commonProps GatewayCommonProps) GatewayConstruct {
	var c GatewayConstruct

	c.Builder = b
	c.Dashboard = commonProps.Dashboard
	c.HandlerAlias = b.setupPubHandler(stack, stackProps)
	c.Gateway = b.setupGateway(stack, c.HandlerAlias)

	b.PublicationTopic.GrantPublish(c.HandlerAlias)

	return c
}

// TODO: the alias does not work with the dashboard - use the actual handler?

func (b GatewayBuilder) setupPubHandler(stack awscdk.Stack, stackProps stackprops.CdkStackProps) awslambda.Alias {
	handlerProps := awslambdago.GoFunctionProps{
		Description:   aws.String("SNS event-raising handler " + stackProps.Version),
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String(b.Entry),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(27)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Tracing:       awslambda.Tracing_ACTIVE,
		Environment:   &b.Environment,
		CurrentVersionOptions: &awslambda.VersionOptions{
			RemovalPolicy: awscdk.RemovalPolicy_RETAIN_ON_UPDATE_OR_DELETE,
		},
	}

	handler := awslambdago.NewGoFunction(stack, aws.String(b.HandlerId), &handlerProps)

	version := handler.CurrentVersion()

	alias := awslambda.NewAlias(stack, aws.String(b.HandlerId+"Alias"), &awslambda.AliasProps{
		AliasName:   aws.String("Live"),
		Description: aws.String("Live version of the PubHandler"),
		Version:     version,
	})

	return alias
}

func (b GatewayBuilder) setupGateway(stack awscdk.Stack, alias awslambda.Alias) awsapigateway.LambdaRestApi {
	stageOptions := awsapigateway.StageOptions{
		StageName:        aws.String(stage),
		LoggingLevel:     awsapigateway.MethodLoggingLevel_INFO,
		TracingEnabled:   aws.Bool(true),
		MetricsEnabled:   aws.Bool(true),
		DataTraceEnabled: aws.Bool(true),
	}

	restApiProps := awsapigateway.LambdaRestApiProps{
		Handler:       alias,
		DeployOptions: &stageOptions,
	}

	return awsapigateway.NewLambdaRestApi(stack, aws.String(b.EndpointId), &restApiProps)
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (c GatewayConstruct) LambdaMetricsGraphWidget() awscloudwatch.GraphWidget {
	region := c.HandlerAlias.Stack().Region()

	invocationsMetric := c.Dashboard.CreateLambdaMetric(*region, "Invocations", c.HandlerAlias.FunctionName(), "Sum")
	errorsMetric := c.Dashboard.CreateLambdaMetric(*region, "Errors", c.HandlerAlias.FunctionName(), "Sum")
	metrics := []awscloudwatch.IMetric{invocationsMetric, errorsMetric}

	return c.Dashboard.CreateGraphWidget(*region, fmt.Sprintf("%s - Invocations & Errors", c.Builder.HandlerId), metrics)
}

func (c GatewayConstruct) GatewayMetricsGraphWidget() awscloudwatch.GraphWidget {
	region := c.HandlerAlias.Stack().Region()

	invocationsMetric := c.Dashboard.CreateGatewayMetric(*region, "Count", c.Builder.EndpointId, stage, "Sum")
	errorsMetric := c.Dashboard.CreateGatewayMetric(*region, "5XXError", c.Builder.EndpointId, stage, "Sum")
	metrics := []awscloudwatch.IMetric{invocationsMetric, errorsMetric}

	return c.Dashboard.CreateGraphWidget(*region, fmt.Sprintf("%s - Invocations & Errors", c.Builder.EndpointId), metrics)
}

func (c GatewayConstruct) TopicMetricsGraphWidget() awscloudwatch.GraphWidget {
	region := c.HandlerAlias.Stack().Region()

	publicationsMetric := c.Dashboard.CreateTopicMetric(*region, "NumberOfMessagesPublished", c.Builder.PublicationTopic.TopicName(), "Sum")
	failsMetric := c.Dashboard.CreateTopicMetric(*region, "NumberOfNotificationsFailed", c.Builder.PublicationTopic.TopicName(), "Sum")
	metrics := []awscloudwatch.IMetric{publicationsMetric, failsMetric}

	return c.Dashboard.CreateGraphWidget(*region, fmt.Sprintf("%s - Publications & Failures", c.Builder.PublicationTopic.Name), metrics)
}
