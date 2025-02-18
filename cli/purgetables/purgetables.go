package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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

	pipe.Command(&b,
		exec.Command("aws", "dynamodb", "scan", "--table-name", tableName, "--attributes-to-get", "PK", "Path"),
		exec.Command("jq", "-c"),
	)

	keys := gjson.Get(b.String(), "Items")

	return keys
}

func purgeTable(tableName string, keys gjson.Result) {
	for _, key := range keys.Array() {
		keyStr := key.String()

		fmt.Printf("deleting %s\n", keyStr)
		_, err := exec.Command("aws", "dynamodb", "delete-item", "--table-name", tableName, "--key", keyStr).Output()
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

	tableIdentifier := os.Args[1]
	names := listTables()

	for _, name := range names {
		if strings.Contains(name, tableIdentifier) {
			fmt.Printf("Purging %s\n", name)
			keys := getKeys(name)

			purgeTable(name, keys)
		}
	}
}
