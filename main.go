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
		fmt.Fprintln(os.Stderr, "usage: mygit <command> [<args>...]")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		initRepo()

	case "cat-file":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: mygit cat-file <hash>")
			os.Exit(1)
		}
		catFile(os.Args[2])

	case "hash-object":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: mygit hash-object <file>")
			os.Exit(1)
		}
		hash := hashObject(os.Args[2])
		fmt.Println(hash)

	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: mygit add <file> [<file>...]")
			os.Exit(1)
		}
		for _, file := range os.Args[2:] {
			addFile(file)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func initRepo() {
	gitDir := ".git"

	if info, err := os.Stat(gitDir); err == nil {
		if !info.IsDir() {
			fmt.Fprintln(os.Stderr, "Error: .git exists but is not a directory")
			os.Exit(1)
		}
		fmt.Println("Reinitialized existing Git directory")
		return
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error checking .git: %s\n", err)
		os.Exit(1)
	}

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
}

func catFile(sha string) {
	path := fmt.Sprintf(".git/objects/%v/%v", sha[0:2], sha[2:])
	fmt.Println("Looking for object at:", path)  // Debug output
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

	parts := strings.SplitN(string(s), "\x00", 2)
	if len(parts) < 2 {
		fmt.Fprintln(os.Stderr, "Error: Corrupt Git object")
		os.Exit(1)
	}
	fmt.Print(parts[1])
}

func hashObject(filePath string) string {
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
		os.Exit(1)
	}

	stats, err := os.Stat(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error stating file: %s\n", err)
		os.Exit(1)
	}

	contentAndHeader := fmt.Sprintf("blob %d\x00%s", stats.Size(), string(file))
	sha := sha1.Sum([]byte(contentAndHeader))
	hash := fmt.Sprintf("%x", sha)
	blobPath := fmt.Sprintf(".git/objects/%s/%s", hash[:2], hash[2:])

	if _, err := os.Stat(blobPath); err == nil {
		return hash
	}

	var buffer bytes.Buffer
	z := zlib.NewWriter(&buffer)
	_, _ = z.Write([]byte(contentAndHeader))
	z.Close()

	os.MkdirAll(filepath.Dir(blobPath), os.ModePerm)
	f, err := os.Create(blobPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %s\n", err)
		os.Exit(1)
	}
	defer f.Close()

	_, err = f.Write(buffer.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to file: %s\n", err)
		os.Exit(1)
	}

	return hash
}

func addFile(filePath string) {
	hash := hashObject(filePath)

	indexPath := ".git/index"
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