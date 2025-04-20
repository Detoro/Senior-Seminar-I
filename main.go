package main

import (
	"fmt"
	"os"
)

type TreeEntry struct {
	mode string
	name string
	sha1 string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: myvcs <command> [<args>...]")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		InitRepo()

	case "content":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: myvcs cat <hash>")
			os.Exit(1)
		}
		CatFile(os.Args[2])

	case "status":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: myvcs staus <hash>")
			os.Exit(1)
		}
		statusFile(os.Args[2])

	case "files":
		if len(os.Args) < 2 {
			fmt.Println("usage: myvcs files")
			os.Exit(1)
		}
		listFilesOnCurrentBranch()

	case "hash":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: myvcs hash <file>")
			os.Exit(1)
		}
		hash := HashObject(os.Args[2])
		fmt.Println(hash)

	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: myvcs add <file> [<file>...]")
			os.Exit(1)
		}
		for _, file := range os.Args[2:] {
			AddFile(file)
		}

	case "read":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: myvcs read <tree-SHA>")
			os.Exit(1)
		}
		ReadTree(os.Args[2])

	case "commit":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: myvcs commit <message>")
			os.Exit(1)
		}
		Commit(os.Args[2])

	case "branch":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: myvcs branch <branch_name>")
			os.Exit(1)
		}
		CreateBranch(os.Args[2])

	case "switch":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: myvcs switch <branch_name>")
			os.Exit(1)
		}
		SwitchBranch(os.Args[2])

	case "clone":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: myvcs clone <url>")
			os.Exit(1)
		}
		err := cloneRepo(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Clone failed: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}
