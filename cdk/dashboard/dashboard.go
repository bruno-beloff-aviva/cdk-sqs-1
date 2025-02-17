package dashboard

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	awslambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/jsii-runtime-go"
)

type Dashboard struct {
	Dashboard awscloudwatch.Dashboard
}

func NewDashboard(stack awscdk.Stack, name string) Dashboard {
	dashboard := awscloudwatch.NewDashboard(stack, jsii.String("eventsDashboard"), &awscloudwatch.DashboardProps{
		DashboardName:   aws.String(name + "-" + *stack.Region()),
		DefaultInterval: awscdk.Duration_Hours(aws.Float64(24)),
	})

	return Dashboard{Dashboard: dashboard}
}

// func (d *Dashboard) AddCloudwatchDashboardMetrics(region string, handler awslambdago.GoFunction) {
// 	invocationsMetric := d.CreateLambdaMetric(region, "Invocations", handler.FunctionName(), "Sum")
// 	errorsMetric := d.CreateLambdaMetric(region, "Errors", handler.FunctionName(), "Sum")

// 	invocationsAndErrors := d.CreateGraphWidget(region, fmt.Sprintf("%s Invocations and Errors", *handler.FunctionName()), []awscloudwatch.IMetric{invocationsMetric, errorsMetric})

// 	row := awscloudwatch.NewRow(invocationsAndErrors)
// 	d.Dashboard.AddWidgets(row)
// }

func (d *Dashboard) AddLambdaMetrics(region string, handler awslambdago.GoFunction, handlerId string) {
	invocationsMetric := d.CreateLambdaMetric(region, "Invocations", handler.FunctionName(), "Sum")
	errorsMetric := d.CreateLambdaMetric(region, "Errors", handler.FunctionName(), "Sum")

	invocationsAndErrors := d.CreateGraphWidget(region, fmt.Sprintf("%s Invocations and Errors", handlerId), []awscloudwatch.IMetric{invocationsMetric, errorsMetric})

	row := awscloudwatch.NewRow(invocationsAndErrors)
	d.Dashboard.AddWidgets(row)
}

func (d *Dashboard) CreateLambdaMetric(region string, metricName string, functionName *string, statistic string) awscloudwatch.IMetric {
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

func (d *Dashboard) CreateCustomMetric(region string, namespace, metricName, SNSName, statistic string) awscloudwatch.IMetric {
	return awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Region:     jsii.String(region),
		Namespace:  jsii.String(namespace),
		MetricName: jsii.String(metricName),
		DimensionsMap: &map[string]*string{
			"SNS": jsii.String(SNSName),
		},
		Period:    awscdk.Duration_Minutes(jsii.Number(5)),
		Statistic: jsii.String(statistic),
	})
}

func (d *Dashboard) CreateGraphWidget(region string, title string, metrics []awscloudwatch.IMetric) awscloudwatch.GraphWidget {
	return awscloudwatch.NewGraphWidget(&awscloudwatch.GraphWidgetProps{
		Region: jsii.String(region),
		Title:  jsii.String(title),
		Left:   &metrics,
		Height: jsii.Number(6),
		Width:  jsii.Number(6),
	})
}

func (d *Dashboard) CreateSingleValueWidget(region string, title string, metrics []awscloudwatch.IMetric) awscloudwatch.SingleValueWidget {
	return awscloudwatch.NewSingleValueWidget(&awscloudwatch.SingleValueWidgetProps{
		Region:               jsii.String(region),
		Title:                jsii.String(title),
		Metrics:              &metrics,
		SetPeriodToTimeRange: jsii.Bool(true),
		Height:               jsii.Number(6),
		Width:                jsii.Number(4),
	})
}
