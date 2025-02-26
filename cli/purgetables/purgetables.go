package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sqstest/service/testreception"
	"strings"

	pipe "github.com/b4b4r07/go-pipe"
	"github.com/tidwall/gjson"
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

func getKeys(tableName string) gjson.Result {
	var b bytes.Buffer

	getArgs := append([]string{"dynamodb", "scan", "--table-name", tableName, "--attributes-to-get"}, testreception.DeletionKeys...)

	err := pipe.Command(&b,
		exec.Command("aws", getArgs...),
		exec.Command("jq", "-c"),
	)
	if err != nil {
		panic(err)
	}

	return gjson.Get(b.String(), "Items")
}

func purgeTable(tableName string, keys gjson.Result) (purgeCount int) {
	for _, key := range keys.Array() {
		keyStr := key.String()

		fmt.Printf("Deleting %s\n", keyStr)
		_, err := exec.Command("aws", "dynamodb", "delete-item", "--table-name", tableName, "--key", keyStr).Output()
		if err != nil {
			panic(err)
		}
		purgeCount++
	}

	return purgeCount
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: purgetables TABLE_PREFIX")
		os.Exit(1)
	}

	tablePrefix := os.Args[1]
	names := listTables()

	purgeCount := 0
	tableCount := 0
	for _, name := range names {
		if strings.HasPrefix(name, tablePrefix) {
			keys := getKeys(name)

			purgeCount += purgeTable(name, keys)
			tableCount++
		}
	}

	fmt.Printf("Purged %d rows(s) in %d table(s).\n", purgeCount, tableCount)
}
