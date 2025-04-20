package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func statusFile(filename string) {
	headPath := ".myvcs/HEAD"
	head, err := os.ReadFile(headPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read HEAD: %s\n", err)
		return
	}

	ref := strings.TrimSpace(strings.TrimPrefix(string(head), "ref: "))
	refPath := filepath.Join(".myvcs", ref)
	commitHashBytes, err := os.ReadFile(refPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read branch ref: %s\n", err)
		return
	}
	commitHash := strings.TrimSpace(string(commitHashBytes))

	commitPath := filepath.Join(".myvcs", "objects", commitHash[:2], commitHash[2:])
	fmt.Println("Looking for object at:", commitPath)
	commitData := readAndInflate(commitPath)
	if commitData == nil {
		fmt.Println("Failed to read commit object")
		return
	}

	parts := strings.SplitN(string(commitData), "\x00", 2)
	if len(parts) < 2 {
		fmt.Println("Invalid commit object format")
		return
	}

	commitBody := parts[1]
	lines := strings.Split(commitBody, "\n")
	var treeHash, parent, author, committer string
	for _, line := range lines {
		if strings.HasPrefix(line, "tree ") {
			treeHash = strings.TrimPrefix(line, "tree ")
		} else if strings.HasPrefix(line, "parent ") {
			parent = strings.TrimPrefix(line, "parent ")
		} else if strings.HasPrefix(line, "author ") {
			author = strings.TrimPrefix(line, "author ")
		} else if strings.HasPrefix(line, "committer ") {
			committer = strings.TrimPrefix(line, "committer ")
		}
	}

	treePath := filepath.Join(".myvcs", "objects", treeHash[:2], treeHash[2:])
	treeData := readAndInflate(treePath)
	if treeData == nil {
		fmt.Println("Failed to read tree object")
		return
	}

	treeParts := strings.SplitN(string(treeData), "\x00", 2)
	if len(treeParts) < 2 {
		fmt.Println("Invalid tree format")
		return
	}

	treeBody := treeParts[1]
	treeBytes := []byte(treeBody)
	found := false
	i := 0
	for i < len(treeBytes) {
		spaceIdx := bytes.IndexByte(treeBytes[i:], ' ')
		i += spaceIdx + 1
		nullIdx := bytes.IndexByte(treeBytes[i:], 0)
		name := treeBytes[i : i+nullIdx]
		i += nullIdx + 1
		i += 20
		if string(name) == filename {
			found = true
			break
		}
	}

	if found {
		fmt.Println("tree", treeHash)
		if parent != "" {
			fmt.Println("parent", parent)
		}
		fmt.Println("author", formatAuthor(author))
		fmt.Println("committer", formatAuthor(committer))
	} else {
		fmt.Printf("%s not found in latest commit\n", filename)
	}
}

func readAndInflate(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	r, err := zlib.NewReader(f)
	if err != nil {
		return nil
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil
	}
	return data
}

func formatAuthor(raw string) string {
	parts := strings.Split(raw, " ")
	if len(parts) < 3 {
		return raw
	}
	timestamp := parts[len(parts)-2]
	t, _ := time.ParseInLocation("20060102150405", timestamp, time.UTC)
	formatted := t.Format("Monday, 02-Jan-06 15:04:05 MST -0700")
	return strings.Join(parts[:len(parts)-2], " ") + " " + formatted
}
