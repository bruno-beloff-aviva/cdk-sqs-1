package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func purgeQueue(url string) {
	var err error

	fmt.Printf("Purging %s\n", url)
	_, err = exec.Command("aws", "sqs", "purge-queue", "--queue-url", url).Output()
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: purgequeues QUEUE_IDENTIFIER")
		os.Exit(1)
	}

	queueIdentifier := os.Args[1]
	urls := listQueues()

	purgeCount := 0
	for _, url := range urls {
		if strings.Contains(url, queueIdentifier) {
			purgeQueue(url)
			purgeCount++
		}
	}

	fmt.Printf("Purged %d queue(s).\n", purgeCount)
}
