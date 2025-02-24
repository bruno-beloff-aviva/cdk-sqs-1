package main

// https://stackoverflow.com/questions/48386497/aws-lambda-list-functions-filter-out-just-function-names

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/b4b4r07/go-pipe"
	"github.com/tidwall/gjson"
)

func listHandlers(region string, _ string) gjson.Result {
	var b bytes.Buffer

	query := "Functions[?starts_with(FunctionName,`SQS`)==`true`].*"

	err := pipe.Command(&b,
		exec.Command("aws", "lambda", "list-functions", "--function-version", "ALL", "--region", region, "--query", query),
		exec.Command("jq", "-c"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", b.String())

	return gjson.Get(b.String(), "Functions") // TODO: get [0] array in array
}

// func getKeys(tableName string) gjson.Result {
// 	var b bytes.Buffer

// 	getArgs := append([]string{"dynamodb", "scan", "--table-name", tableName, "--attributes-to-get"}, testreception.DeletionKeys...)

// 	err := pipe.Command(&b,
// 		exec.Command("aws", getArgs...),
// 		exec.Command("jq", "-c"),
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return gjson.Get(b.String(), "Items")
// }

// func purgeTable(tableName string, keys gjson.Result) {
// 	for _, key := range keys.Array() {
// 		keyStr := key.String()

// 		fmt.Printf("deleting %s\n", keyStr)
// 		_, err := exec.Command("aws", "dynamodb", "delete-item", "--table-name", tableName, "--key", keyStr).Output()
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// }

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: purgetable REGION STACK_IDENTIFIER")
		os.Exit(1)
	}

	region := os.Args[1]
	stackIdentifier := os.Args[2]

	handlers := listHandlers(region, stackIdentifier)
	fmt.Printf("handlers %s\n", handlers)

	for _, handler := range handlers.Array() {
		if strings.HasPrefix(handler.Get("FunctionName").String(), stackIdentifier) {
			fmt.Printf("Purging %s\n", handler)
			fmt.Println("-")

			// TODO: get the right properties in order to find old versions
		}
	}
}
