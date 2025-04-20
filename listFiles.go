package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func listFilesOnCurrentBranch() {
	branch, err := getCurrentBranch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read current branch: %s\n", err)
		return
	}
	fmt.Println("Current branch:", branch)

	branchPath := filepath.Join(".myvcs", "refs", "heads", branch)
	commitHashBytes, err := os.ReadFile(branchPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read branch ref: %s\n", err)
		return
	}
	commitHash := strings.TrimSpace(string(commitHashBytes))

	visited := make(map[string]bool)
	files := make(map[string]bool)

	for commitHash != "" && !visited[commitHash] {
		visited[commitHash] = true

		commitPath := filepath.Join(".myvcs", "objects", commitHash[:2], commitHash[2:])
		commitData := readAndInflate(commitPath)
		if commitData == nil {
			fmt.Println("Failed to read commit object")
			return
		}
		parts := strings.SplitN(string(commitData), "\x00", 2)
		if len(parts) < 2 {
			fmt.Println("Invalid commit format")
			return
		}
		commitBody := parts[1]
		lines := strings.Split(commitBody, "\n")
		var treeHash, parent string
		for _, line := range lines {
			if strings.HasPrefix(line, "tree ") {
				treeHash = strings.TrimPrefix(line, "tree ")
			} else if strings.HasPrefix(line, "parent ") {
				parent = strings.TrimPrefix(line, "parent ")
			}
		}

		collectTreeFiles(treeHash, files, "")
		commitHash = parent
	}

	for name := range files {
		fmt.Println(name)
	}
}

func collectTreeFiles(treeHash string, files map[string]bool, prefix string) {
	treePath := filepath.Join(".myvcs", "objects", treeHash[:2], treeHash[2:])
	treeData := readAndInflate(treePath)
	if treeData == nil {
		return
	}

	nullIdx := bytes.IndexByte(treeData, 0)
	if nullIdx == -1 || nullIdx+1 >= len(treeData) {
		return
	}
	treeBytes := treeData[nullIdx+1:]

	i := 0
	for i < len(treeBytes) {
		spaceIdx := bytes.IndexByte(treeBytes[i:], ' ')
		if spaceIdx == -1 {
			break
		}
		mode := treeBytes[i : i+spaceIdx]
		i += spaceIdx + 1

		nullIdx := bytes.IndexByte(treeBytes[i:], 0)
		if nullIdx == -1 {
			break
		}
		name := treeBytes[i : i+nullIdx]
		i += nullIdx + 1

		if i+20 > len(treeBytes) {
			break
		}
		sha := fmt.Sprintf("%x", treeBytes[i:i+20])
		i += 20

		fullName := filepath.Join(prefix, string(name))

		if string(mode) == "40000" {
			// recurse into subdirectory
			collectTreeFiles(sha, files, fullName)
		} else {
			files[fullName] = true
		}
	}
}



func getCurrentBranch() (string, error) {
	headPath := ".myvcs/HEAD"
	headData, err := os.ReadFile(headPath)
	if err != nil {
		return "", err
	}
	head := strings.TrimSpace(string(headData))
	if strings.HasPrefix(head, "ref: ") {
		parts := strings.Split(head, "/")
		return parts[len(parts)-1], nil
	}
	return head, nil
}