package main

import (
	"fmt"
	"sqstest/apiclient"
)

func main() {
	client := apiclient.NewClient("https://j7zugt2z5d.execute-api.eu-west-2.amazonaws.com/prod/")

	test := "1"
	function := "ok" // TODO: enum

	response := client.Get(test, function)

	fmt.Printf("response: %s\n", response)
}
