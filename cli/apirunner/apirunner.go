package main

import (
	"fmt"
	"os"
	"sqstest/apiclient"
)

const interval = 2

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: apirunner PUB_API_URL")
		os.Exit(1)
	}

	baseUrl := os.Args[1]
	client := apiclient.NewClient(baseUrl, interval)

	tape := apiclient.NewTape()

	tape.Add("1", "ok1", 10)
	tape.Add("1", "suspend", 1)
	tape.Add("1", "ok2", 10)
	tape.Add("1", "resume", 1)
	tape.Add("1", "ok3", 10)

	client.Run(tape)
}
