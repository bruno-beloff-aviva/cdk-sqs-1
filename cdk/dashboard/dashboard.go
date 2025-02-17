package dashboard

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	awslambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/jsii-runtime-go"
)

// BuilderProps groups all configuration needed to build an event handler
type BuilderProps struct {
	CommsTable                   awsdynamodb.ITable
	StandardEnvironmentVariables map[string]*string
	StandardTimeout              awscdk.Duration
	MetricNamespace              string
	AlertTopics                  []awssns.ITopic
	QueueKey                     awskms.IKey
	EntryQueue                   awssqs.IQueue
	EntryHandler                 awslambdago.GoFunction
	CloudwatchDashboard          awscloudwatch.Dashboard
}

// Builder handles the configuration and creation of event handlers
type Builder struct {
	eventName   string
	entry       string
	environment map[string]*string
}

// NewBuilder begins the configuration of a new EventHandler
func NewBuilder(eventName string, entry string) *Builder {
	return &Builder{
		eventName:   eventName,
		entry:       entry,
		environment: map[string]*string{},
	}
}

func (b *Builder) AddCloudwatchDashboardMetrics(region string, props BuilderProps, handler awslambdago.GoFunction) {
	invocationsMetric := b.CreateLambdaMetric(region, "Invocations", handler.FunctionName(), "Sum")
	errorsMetric := b.CreateLambdaMetric(region, "Errors", handler.FunctionName(), "Sum")

	invocationsAndErrors := b.CreateGraphWidget(region, fmt.Sprintf("%s Invocations and Errors", b.eventName), []awscloudwatch.IMetric{invocationsMetric, errorsMetric})

	row := awscloudwatch.NewRow(invocationsAndErrors)
	props.CloudwatchDashboard.AddWidgets(row)
}

func (b *Builder) CreateLambdaMetric(region string, metricName string, functionName *string, statistic string) awscloudwatch.IMetric {
	return awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Region:     jsii.String(region),
		Namespace:  jsii.String("AWS/Lambda"),
		MetricName: jsii.String(metricName),
		DimensionsMap: &map[string]*string{
			"FunctionName": functionName,
		},
		Period:    awscdk.Duration_Minutes(jsii.Number(5)),
		Statistic: jsii.String(statistic),
	})
}

func (b *Builder) CreateCustomMetric(region string, namespace, metricName, eventName, statistic string) awscloudwatch.IMetric {
	return awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Region:     jsii.String(region),
		Namespace:  jsii.String(namespace),
		MetricName: jsii.String(metricName),
		DimensionsMap: &map[string]*string{
			"event": jsii.String(eventName),
		},
		Period:    awscdk.Duration_Minutes(jsii.Number(5)),
		Statistic: jsii.String(statistic),
	})
}

func (b *Builder) CreateGraphWidget(region string, title string, metrics []awscloudwatch.IMetric) awscloudwatch.GraphWidget {
	return awscloudwatch.NewGraphWidget(&awscloudwatch.GraphWidgetProps{
		Region: jsii.String(region),
		Title:  jsii.String(title),
		Left:   &metrics,
		Height: jsii.Number(6),
		Width:  jsii.Number(6),
	})
}

func (b *Builder) CreateSingleValueWidget(region string, title string, metrics []awscloudwatch.IMetric) awscloudwatch.SingleValueWidget {
	return awscloudwatch.NewSingleValueWidget(&awscloudwatch.SingleValueWidgetProps{
		Region:               jsii.String(region),
		Title:                jsii.String(title),
		Metrics:              &metrics,
		SetPeriodToTimeRange: jsii.Bool(true),
		Height:               jsii.Number(6),
		Width:                jsii.Number(4),
	})
}
