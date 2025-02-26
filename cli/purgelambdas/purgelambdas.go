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

func listLambdaGroups(stackName string) (lambdaGroups map[string][]Lambda) {
	var b bytes.Buffer

	query := fmt.Sprintf("Functions[?starts_with(FunctionName,`%s`)==`true`].*", stackName)

	err := pipe.Command(&b,
		// list-functions returns in version order, with $latest first...
		exec.Command("aws", "lambda", "list-functions", "--function-version", "ALL", "--query", query),
		exec.Command("jq", "-c"),
	)
	if err != nil {
		panic(err)
	}

	lambdaGroups = make(map[string][]Lambda)

	for _, resource := range gjson.Parse(b.String()).Array() {
		lambda := NewLambda(resource.Array())
		lambdaGroups[lambda.PhysicalID] = append(lambdaGroups[lambda.PhysicalID], lambda)
	}

	return lambdaGroups
}

func purgeGroup(group []Lambda, keepCount int) (purgeCount int) {
	for i := 1; i < len(group)-keepCount; i++ {
		lambda := group[i]
		fmt.Printf("Purging %s\n", lambda.FunctionArn)

		_, err := exec.Command("aws", "lambda", "delete-function", "--function-name", lambda.PhysicalID, "--qualifier", lambda.Version()).Output()
		if err != nil {
			panic(err)
		}
		purgeCount++
	}

	return purgeCount
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: purgelambdas STACK_NAME")
		os.Exit(1)
	}

	keepCount := 2
	stackName := os.Args[1]

	purgeCount := 0
	for _, group := range listLambdaGroups(stackName) {
		purgeCount += purgeGroup(group, keepCount)
	}

	fmt.Printf("Purged %d lambda(s).\n", purgeCount)
}
