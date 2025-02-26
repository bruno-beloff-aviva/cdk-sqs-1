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

type Lambda struct {
	PhysicalID  string
	FunctionArn string
	Runtime     string
}

func NewLambda(jdict []gjson.Result) Lambda {
	return Lambda{
		PhysicalID:  jdict[0].String(),
		FunctionArn: jdict[1].String(),
		Runtime:     jdict[2].String(),
	}
}

func (l Lambda) String() string {
	return fmt.Sprintf("Lambda:{PhysicalID: %s, FunctionArn: %s, Runtime: %s}", l.PhysicalID, l.FunctionArn, l.Runtime)
}

func (l Lambda) Version() string {
	return strings.Split(l.FunctionArn, ":")[7]
}

func listLambdas(stackName string) (lambdas []Lambda) {
	var b bytes.Buffer

	query := fmt.Sprintf("Functions[?starts_with(FunctionName,`%s`)==`true`].*", stackName)

	err := pipe.Command(&b,
		exec.Command("aws", "lambda", "list-functions", "--function-version", "ALL", "--query", query),
		exec.Command("jq", "-c"),
	)
	if err != nil {
		panic(err)
	}

	for _, resource := range gjson.Parse(b.String()).Array() {
		lambdas = append(lambdas, NewLambda(resource.Array()))
	}

	return lambdas
}

func purgeGroup(group []Lambda, keepCount int) {
	for i := 1; i < len(group)-keepCount; i++ {
		lambda := group[i]
		fmt.Printf("Purging %s\n", lambda)

		_, err := exec.Command("aws", "lambda", "delete-function", "--function-name", lambda.PhysicalID, "--qualifier", lambda.Version()).Output()
		if err != nil {
			fmt.Println("error: ", err)
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: purgetable STACK_NAME")
		os.Exit(1)
	}

	keepCount := 2
	stackName := os.Args[1]
	lambdaGroups := make(map[string][]Lambda)

	for _, lambda := range listLambdas(stackName) {
		lambdaGroups[lambda.PhysicalID] = append(lambdaGroups[lambda.PhysicalID], lambda)
	}

	for _, group := range lambdaGroups {
		purgeGroup(group, keepCount)
	}
}
