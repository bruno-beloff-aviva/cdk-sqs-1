package main

// https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html#golang-handler-signatures
// https://stackoverflow.com/questions/37365009/how-to-invoke-an-aws-lambda-function-asynchronously
// proper logging: https://github.com/awsdocs/aws-lambda-developer-guide/blob/main/sample-apps/blank-go/function/main.go

import (
	"context"
	"os"
	"sqstest/lambda/handler/subhandler"
	"sqstest/manager/dbmanager"
	"sqstest/service/sub"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/joerdav/zapray"
)

func main() {
	logger, err1 := zapray.NewDevelopment() // log level is set using this: NewProduction(), NewDevelopment(), NewExample()

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
	dbManager := dbmanager.NewDynamoManager(logger, cfg, tableName)
	tableIsAvailable := dbManager.TableIsAvailable(ctx)

	if !tableIsAvailable {
		panic("Table not available: " + tableName)
	}

	//	service...
	subService := sub.NewSuspendableService(logger, cfg, dbManager, "suspendable")

	//	lambda...
	subHandler := subhandler.NewSubHandler(logger, subService)

	lambda.StartWithOptions(subHandler.Handle, lambda.WithEnableSIGTERM(func() {
		logger.Info("<<< Lambda container shutting down.")
	}))
}
