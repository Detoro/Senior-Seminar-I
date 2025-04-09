package main

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strings"
)

func CatFile(sha string) {
	path := fmt.Sprintf(".myvcs/objects/%v/%v", sha[0:2], sha[2:])
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

	parts := strings.Split(string(s), "\x00")
	if len(parts) < 2 {
		fmt.Fprintf(os.Stderr, "Invalid blob format\n")
		os.Exit(1)
	}
	fmt.Print(parts[1])
}