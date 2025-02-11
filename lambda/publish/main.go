// https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html#golang-handler-signatures
// https://stackoverflow.com/questions/37365009/how-to-invoke-an-aws-lambda-function-asynchronously
// proper logging: https://github.com/awsdocs/aws-lambda-developer-guide/blob/main/sample-apps/blank-go/function/main.go

package main

import (
	"context"
	"os"
	"sqstest/lambda/publish/publisher"
	"sqstest/service"
	"sqstest/sqsmanager"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/joerdav/zapray"
)

func main() {
	logger, err1 := zapray.NewDevelopment() // log level is set using this: NewProduction(), NewDevelopment(), NewExample()

	if err1 != nil {
		panic("failed to create logger: " + err1.Error())
	}
	logger.Info("*** main")

	//	context...
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		panic("err: " + err.Error())
	}

	//	environment...
	queueUrl := os.Getenv("QUEUE_URL")
	logger.Info("queueUrl: " + queueUrl)

	//	managers...
	sqsManager := sqsmanager.NewSQSManager(logger, cfg)

	//	services...
	publishService := service.NewPublishService(logger, sqsManager, queueUrl)

	//	lambda...
	publishHandler := publisher.NewPublishHandler(logger, publishService)

	lambda.StartWithOptions(publishHandler.Handle, lambda.WithEnableSIGTERM(func() {
		logger.Info("Lambda container shutting down.")
	}))
}
