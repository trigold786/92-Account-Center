package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run generate_down.go <migration_file.sql>")
		fmt.Println("Generates a _down.sql file for the given up migration.")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	content, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", inputFile, err)
		os.Exit(1)
	}

	tables := extractTables(string(content))
	if len(tables) == 0 {
		fmt.Println("No CREATE TABLE statements found.")
		return
	}

	dir := filepath.Dir(inputFile)
	base := filepath.Base(inputFile)
	ext := filepath.Ext(base)
	nameOnly := strings.TrimSuffix(base, ext)

	outputFile := filepath.Join(dir, nameOnly+"_down"+ext)

	var sb strings.Builder
	sb.WriteString("-- +goose Down\n")
	for _, table := range tables {
		sb.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;\n", table))
	}

	if err := os.WriteFile(outputFile, []byte(sb.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", outputFile, err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s with %d table(s)\n", outputFile, len(tables))
}

func extractTables(content string) []string {
	var tables []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		upper := strings.ToUpper(trimmed)
		if strings.HasPrefix(upper, "CREATE TABLE ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				tableName := strings.TrimPrefix(parts[2], "IF NOT EXISTS ")
				tableName = strings.TrimRight(tableName, " (")
				tables = append(tables, tableName)
			}
		}
	}
	return tables
}
