package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		gitDir := ".git"
	
		// Check if .git exists and whether it's a directory
		if info, err := os.Stat(gitDir); err == nil {
			if !info.IsDir() {
				fmt.Fprintf(os.Stderr, "Error: .git exists but is not a directory\n")
				os.Exit(1)
			}
			fmt.Println("Reinitialized existing Git directory")
		} else if os.IsNotExist(err) {
			// Create the necessary directories
			for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating directory %s: %s\n", dir, err)
					os.Exit(1)
				}
			}
	
			headFileContents := []byte("ref: refs/heads/main\n")
			if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing .git/HEAD: %s\n", err)
				os.Exit(1)
			}
	
			fmt.Println("Initialized empty Git repository")
		} else {
			fmt.Fprintf(os.Stderr, "Error checking .git: %s\n", err)
			os.Exit(1)
		}
	

	case "cat-file":
		sha := os.Args[3]
		path := fmt.Sprintf(".git/objects/%v/%v", sha[0:2], sha[2:])
		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %s\n", err)
			os.Exit(1)
		}
		defer file.Close()

		r, err := zlib.NewReader(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating zlib reader: %s\n", err)
			os.Exit(1)
		}
		defer r.Close()

		s, err := io.ReadAll(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
			os.Exit(1)
		}

		parts := strings.Split(string(s), "\x00")
		fmt.Print(parts[1])
		r.Close()

	case "hash-object":
		file, _ := os.ReadFile(os.Args[3])
		stats, _ := os.Stat(os.Args[3])
		content := string(file)
		contentAndHeader := fmt.Sprintf("blob %d\x00%s", stats.Size(), content)
		sha := (sha1.Sum([]byte(contentAndHeader)))
		hash := fmt.Sprintf("%x", sha)
		blobPath := fmt.Sprintf(".git/objects/%s/%s", hash[:2], hash[2:])

		var buffer bytes.Buffer
		z := zlib.NewWriter(&buffer)
		z.Write([]byte(contentAndHeader))
		z.Close()
		os.MkdirAll(filepath.Dir(blobPath), os.ModePerm)
		f, err := os.Create(blobPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %s\n", err)
			os.Exit(1)
		}
		defer f.Close()

		f.Write(buffer.Bytes())
		_, err = f.Write(buffer.Bytes())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %s\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
