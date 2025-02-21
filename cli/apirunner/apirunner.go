package main

import (
	"sqstest/apiclient"
)

// TODO: present the URL as an argument

const baseUrl = "https://fsfv22wx52.execute-api.eu-west-2.amazonaws.com/prod/"
const interval = 2

func main() {
	client := apiclient.NewClient(baseUrl, interval)

	tape := apiclient.NewTape()

	tape.Add("1", "ok1", 10)
	tape.Add("1", "suspend", 1)
	tape.Add("1", "ok2", 10)
	tape.Add("1", "resume", 1)
	tape.Add("1", "ok3", 10)

	client.Run(tape)
}
