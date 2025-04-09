package main

import (
	"fmt"
	"os"
	"strings"
)

func AddFile(filePath string) {
	hash := HashObject(filePath)

	indexPath := ".myvcs/index"
	entry := fmt.Sprintf("%s %s\n", hash, filePath)

	var indexContent string
	if data, err := os.ReadFile(indexPath); err == nil {
		indexContent = string(data)
	}

	if !strings.Contains(indexContent, filePath) {
		indexContent += entry
		if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating index: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added %s\n", filePath)
	} else {
		fmt.Printf("%s is already staged\n", filePath)
	}
}