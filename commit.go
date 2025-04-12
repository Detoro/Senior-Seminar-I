package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CreateCommit(treeSHA string, parentSHA string, author string, committer string, message string) string {
	timestamp := time.Now().Format(time.RFC850)
	var content bytes.Buffer
	content.WriteString(fmt.Sprintf("tree %s\n", treeSHA))
	if parentSHA != "" {
		content.WriteString(fmt.Sprintf("parent %s\n", parentSHA))
	}
	content.WriteString(fmt.Sprintf("author %s %s +0000\n", author, timestamp)) //FIXME: Timezone
	content.WriteString(fmt.Sprintf("committer %s %s +0000\n", committer, timestamp)) //FIXME: Timezone
	content.WriteString(fmt.Sprintf("\n%s\n", message))

	header := fmt.Sprintf("commit %d\x00", content.Len())
	fullContent := header + content.String()
	sha1 := fmt.Sprintf("%x", sha1.Sum([]byte(fullContent)))
	objectPath := fmt.Sprintf(".myvcs/objects/%s/%s", sha1[:2], sha1[2:])

	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write([]byte(fullContent))
	w.Close()

	os.MkdirAll(filepath.Dir(objectPath), 0755)
	if err := os.WriteFile(objectPath, compressed.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing commit object: %s\n", err)
		os.Exit(1)
	}
	return sha1
}

func Commit(message string) {
	// 1. Create tree
	entries, err := WriteTree(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing tree: %s\n", err)
		os.Exit(1)
	}
	treeSHA := CreateTree(entries)

	// 2. Get parent commit
	parentSHA := ""
	headRef, err := os.ReadFile(".myvcs/HEAD")
	if err == nil {
		branchName := strings.TrimSpace(strings.TrimPrefix(string(headRef), "ref: refs/heads/"))
		headPath := fmt.Sprintf(".myvcs/refs/heads/%s", branchName)
		parentBytes, err := os.ReadFile(headPath)
		if err == nil {
			parentSHA = strings.TrimSpace(string(parentBytes))
		}
	}

	// 3. Create commit object
	author := "Adetoro <awakinola16@my.fisk.edu>" //FIXME: Get from config
	committer := "Adetoro <awakinola16@my.fisk.edu>" //FIXME: Get from config
	commitSHA := CreateCommit(treeSHA, parentSHA, author, committer, message)

	// 4. Update HEAD
	headRef, err = os.ReadFile(".myvcs/HEAD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading HEAD: %s\n", err)
		os.Exit(1)
	}
	headContent := strings.TrimSpace(string(headRef)) //  Clean up the content

	var headPath string
	if strings.HasPrefix(headContent, "ref: refs/heads/") {
		branchName := strings.TrimPrefix(headContent, "ref: refs/heads/")
		headPath = fmt.Sprintf(".myvcs/refs/heads/%s", branchName)
	} else {
		//  Detached HEAD - write the commit SHA directly to HEAD
		headPath = ".myvcs/HEAD"
	}

	if err := os.WriteFile(headPath, []byte(commitSHA+"\n"), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating %s: %s\n", headPath, err)
		os.Exit(1)
	}
}