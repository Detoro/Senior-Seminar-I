package main

import (
	"fmt"
	"os"
)

func InitRepo() {
	gitDir := ".myvcs"

	if info, err := os.Stat(gitDir); err == nil {
		if !info.IsDir() {
			fmt.Fprintln(os.Stderr, "Error: .myvcs exists but is not a directory")
			os.Exit(1)
		}
		fmt.Println("Reinitialized existing VCS directory")
		return
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error checking .myvcs: %s\n", err)
		os.Exit(1)
	}

	for _, dir := range []string{".myvcs", ".myvcs/objects", ".myvcs/refs", ".myvcs/refs/heads"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %s\n", dir, err)
			os.Exit(1)
		}
	}

	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".myvcs/HEAD", headFileContents, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing .myvcs/HEAD: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Initialized empty myvcs repository")
}
