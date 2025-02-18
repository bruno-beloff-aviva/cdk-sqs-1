package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func listTables() []string {
	var out []byte
	var listing map[string][]string
	var err error

	out, err = exec.Command("aws", "dynamodb", "list-tables").Output()
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(out, &listing)
	if err != nil {
		panic(err)
	}

	return listing["TableNames"]
}

func getKeys(tableName string) []map[string]string {
	var out []byte
	var listing map[string]any
	var err error

	out, err = exec.Command("aws", "dynamodb", "scan", "--table-name", tableName, "--attributes-to-get", "PK", "Path").Output()
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(out, &listing)
	if err != nil {
		panic(err)
	}

	items := listing["Items"].([]interface{})
	var keys []map[string]string

	for _, item := range items {
		pk := item.(map[string]interface{})["PK"]
		pkValue := pk.(map[string]interface{})["S"]

		path := item.(map[string]interface{})["Path"]
		pathValue := path.(map[string]interface{})["S"]

		keys = append(keys, map[string]string{"PK": pkValue.(string), "Path": pathValue.(string)})
	}
	return keys
}

func purgeTable(tableName string, keys []map[string]string) {
	for _, key := range keys {
		keySpec := map[string]string{"PK": key["PK"], "Path": key["Path"]}
		keyJson, _ := json.Marshal(keySpec)

		fmt.Printf("deleting %s\n", keySpec)

		_, err := exec.Command("aws", "dynamodb", "delete-item", "--table-name", tableName, "--key", string(keyJson)).Output()
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: purgetable TABLE_IDENTIFIER")
		os.Exit(1)
	}

	queueIdentifier := os.Args[1]
	names := listTables()

	for _, name := range names {
		if strings.Contains(name, queueIdentifier) {
			fmt.Printf("Purging %s\n", name)
			keys := getKeys(name)

			purgeTable(name, keys)
		}
	}
}
