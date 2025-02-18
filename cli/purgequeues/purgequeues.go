package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func listQueues() []string {
	var out []byte
	var listing map[string][]string
	var err error

	out, err = exec.Command("aws", "sqs", "list-queues").Output()
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(out, &listing)
	if err != nil {
		panic(err)
	}

	return listing["QueueUrls"]
}

func purgeQueue(queueUrl string) {
	var err error

	_, err = exec.Command("aws", "sqs", "purge-queue", "--queue-url", queueUrl).Output()
	if err != nil {
		panic(err)
	}
}

func main() {
	urls := listQueues()

	for _, url := range urls {
		if strings.Contains(url, "SQS1Stack-TestQueue") {
			fmt.Printf("Purging %s\n", url)
			purgeQueue(url)
		}
	}
}
