package main

// https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html#golang-handler-signatures
// https://stackoverflow.com/questions/37365009/how-to-invoke-an-aws-lambda-function-asynchronously
// proper logging: https://github.com/awsdocs/aws-lambda-developer-guide/blob/main/sample-apps/blank-go/function/main.go

import (
	"context"
	"os"
	"sqstest/lambda/handler/pubhandler"
	"sqstest/service/pub"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/joerdav/zapray"
)

func main() {
	logger, err1 := zapray.NewProduction() // log level is set using this: NewProduction(), NewDevelopment(), NewExample()

	if err1 != nil {
		panic("failed to create logger: " + err1.Error())
	}
	logger.Info(">>> publish main")

	//	context...
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		panic("err: " + err.Error())
	}

	//	environment...
	version := os.Getenv("VERSION")
	logger.Info("version: " + version)

	topicArn := os.Getenv("TOPIC_ARN")
	logger.Info("topicArn: " + topicArn)

	//	service...
	pubService := pub.NewSNSPubService(logger, cfg, topicArn)

	//	lambda...
	publishHandler := pubhandler.NewPubHandler(logger, pubService)

	lambda.StartWithOptions(publishHandler.Handle, lambda.WithEnableSIGTERM(func() {
		logger.Info("<<< Lambda container shutting down.")
	}))
}
