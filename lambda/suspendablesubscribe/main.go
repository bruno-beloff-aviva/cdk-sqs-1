// https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html#golang-handler-signatures
// https://stackoverflow.com/questions/37365009/how-to-invoke-an-aws-lambda-function-asynchronously
// proper logging: https://github.com/awsdocs/aws-lambda-developer-guide/blob/main/sample-apps/blank-go/function/main.go

package main

import (
	"context"
	"os"
	"sqstest/dynamomanager"
	"sqstest/lambda/subscriber"
	"sqstest/service"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/joerdav/zapray"
)

func main() {
	logger, err1 := zapray.NewProduction() // log level is set using this: NewProduction(), NewDevelopment(), NewExample()

	if err1 != nil {
		panic("failed to create logger: " + err1.Error())
	}
	logger.Info(">>> subscribe main")

	//	context...
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		panic("err: " + err.Error())
	}

	//	environment...
	version := os.Getenv("VERSION")
	logger.Info("version: " + version)

	tableName := os.Getenv("MESSAGE_TABLE_NAME")
	logger.Info("tableName: " + tableName)

	suspended := os.Getenv("SUSPENDED") == "true"
	logger.Info("suspended: " + strconv.FormatBool(suspended))

	//	managers...
	dbManager := dynamomanager.NewDynamoManager(logger, cfg, tableName)
	tableIsAvailable := dbManager.TableIsAvailable(ctx)

	if !tableIsAvailable {
		panic("Table not available: " + tableName)
	}

	//	service...
	subscribeService := service.NewSuspendableService(logger, cfg, dbManager)

	//	lambda...
	subscribeHandler := subscriber.NewSubscribeHandler(logger, *subscribeService)

	lambda.StartWithOptions(subscribeHandler.Handle, lambda.WithEnableSIGTERM(func() {
		logger.Info("<<< Lambda container shutting down.")
	}))
}
