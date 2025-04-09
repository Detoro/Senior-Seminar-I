package main

import (
	"fmt"
	"os"
	"strings"
)

func CreateBranch(branchName string) {
	// 1. Get current HEAD commit
	headRef, err := os.ReadFile(".myvcs/HEAD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading HEAD: %s\n", err)
		os.Exit(1)
	}
	currentBranchName := strings.TrimSpace(strings.TrimPrefix(string(headRef), "ref: refs/heads/"))
	headPath := fmt.Sprintf(".myvcs/refs/heads/%s", currentBranchName)
	commitSHA, err := os.ReadFile(headPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %s\n", headPath, err)
		os.Exit(1)
	}
	// 2. Create new branch file
	newBranchPath := fmt.Sprintf(".myvcs/refs/heads/%s", branchName)
	if _, err := os.Stat(newBranchPath); err == nil {
		fmt.Fprintf(os.Stderr, "Branch %s already exists\n", branchName)
		os.Exit(1)
	}
	if err := os.WriteFile(newBranchPath, commitSHA, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating branch %s: %s\n", branchName, err)
		os.Exit(1)
	}

	fmt.Printf("Created branch %s\n", branchName)
}

func SwitchBranch(branchName string) {
	branchPath := fmt.Sprintf(".myvcs/refs/heads/%s", branchName)
	if _, err := os.Stat(branchPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Branch %s does not exist\n", branchName)
		os.Exit(1)
	}
	if err := os.WriteFile(".myvcs/HEAD", []byte(fmt.Sprintf("ref: refs/heads/%s\n", branchName)), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error switching to branch %s: %s\n", branchName, err)
		os.Exit(1)
	}
	fmt.Printf("Switched to branch %s\n", branchName)
}