package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func listBuckets() (bucketNames []string) {
	var out []byte
	var err error

	out, err = exec.Command("aws", "s3", "ls").Output()
	if err != nil {
		log.Fatal(err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		tokens := strings.Split(line, " ")
		if len(tokens) < 3 {
			continue
		}
		bucketNames = append(bucketNames, tokens[2])
	}

	return bucketNames
}

func deleteBucket(bucketName string) (success bool) {
	fmt.Printf("Purging %s\n", bucketName)
	_, err := exec.Command("aws", "s3", "rb", "s3://"+bucketName, "--force").Output()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return false
	}

	return true
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: buckets BUCKET_IDENTIFIER")
		os.Exit(1)
	}

	bucketIdentifier := os.Args[1]

	foundCount := 0
	purgeCount := 0
	for _, bucketName := range listBuckets() {
		if strings.Contains(bucketName, bucketIdentifier) {
			ok := deleteBucket(bucketName)
			if ok {
				purgeCount++
			}
			foundCount++
		}
	}

	fmt.Printf("Found %d buckets(s), purged %d.\n", foundCount, purgeCount)
}
