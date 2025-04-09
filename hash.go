package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
)

func HashObject(filePath string) string {
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
	blobPath := fmt.Sprintf(".myvcs/objects/%s/%s", hash[:2], hash[2:])

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

