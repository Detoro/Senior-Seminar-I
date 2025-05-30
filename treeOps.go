package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"encoding/hex"
)

func ReadTree(filepath string) {

treeSHA := filepath
treePath := fmt.Sprintf(".myvcs/objects/%s/%s", treeSHA[:2], treeSHA[2:])

file, err := os.Open(treePath)
if err != nil {
	fmt.Fprintf(os.Stderr, "Error opening tree object: %s\n", err)
	os.Exit(1)
}
defer file.Close()

// Decompress using zlib
r, err := zlib.NewReader(file)
if err != nil {
	fmt.Fprintf(os.Stderr, "Error decompressing tree object: %s\n", err)
	os.Exit(1)
}
defer r.Close()

// Read decompressed data
data, err := io.ReadAll(r)
if err != nil {
	fmt.Fprintf(os.Stderr, "Error reading tree object: %s\n", err)
	os.Exit(1)
}

// Skip the "tree <size>\x00" header
nullIndex := bytes.IndexByte(data, 0)
if nullIndex == -1 {
	fmt.Fprintf(os.Stderr, "Invalid tree object format\n")
	os.Exit(1)
}
data = data[nullIndex+1:]

// Parse tree entries
var i int
for i < len(data) {
	// Extract file mode
	endOfMode := bytes.IndexByte(data[i:], ' ')
	if endOfMode == -1 {
		fmt.Fprintf(os.Stderr, "Invalid tree entry format\n")
		os.Exit(1)
	}
	mode := string(data[i : i+endOfMode])
	i += endOfMode + 1

	// Extract filename
	endOfFilename := bytes.IndexByte(data[i:], 0)
	if endOfFilename == -1 {
		fmt.Fprintf(os.Stderr, "Invalid tree entry format\n")
		os.Exit(1)
	}
	filename := string(data[i : i+endOfFilename])
	i += endOfFilename + 1

	// Extract SHA-1 (20 bytes)
	if i+20 > len(data) {
		fmt.Fprintf(os.Stderr, "Invalid SHA-1 length in tree object\n")
		os.Exit(1)
	}
	objectSHA := fmt.Sprintf("%x", data[i:i+20])
	i += 20

	// Print the tree entry
	fmt.Printf("%s %s %s\n", mode, objectSHA, filename)
}
}

func CreateTree(entries []TreeEntry) string {
var content bytes.Buffer
// ... inside CreateTree
for _, entry := range entries {
    // Get the raw 20-byte SHA-1 hash from the hex string
    sha1Bytes, err := hex.DecodeString(entry.sha1)
    if err != nil {
         // Handle error: invalid SHA1 string
         fmt.Fprintf(os.Stderr, "Error decoding SHA-1 hex string %s: %s\n", entry.sha1, err)
         os.Exit(1) // Or return error
    }
    if len(sha1Bytes) != 20 {
         // Handle error: incorrect SHA-1 byte length
         fmt.Fprintf(os.Stderr, "Error: Decoded SHA-1 has incorrect byte length: %d\n", len(sha1Bytes))
         os.Exit(1) // Or return error
    }

    content.WriteString(fmt.Sprintf("%s %s\x00", entry.mode, entry.name))
    content.Write(sha1Bytes) // Write the raw bytes
}

header := fmt.Sprintf("tree %d\x00", content.Len())
fullContent := header + content.String()

sha1 := fmt.Sprintf("%x", sha1.Sum([]byte(fullContent)))
objectPath := fmt.Sprintf(".myvcs/objects/%s/%s", sha1[:2], sha1[2:])

if _, err := os.Stat(objectPath); err == nil {
	return sha1
}

var compressed bytes.Buffer
w := zlib.NewWriter(&compressed)
w.Write([]byte(fullContent))
w.Close()

os.MkdirAll(filepath.Dir(objectPath), 0755)
if err := os.WriteFile(objectPath, compressed.Bytes(), 0644); err != nil {
	fmt.Fprintf(os.Stderr, "Error writing tree object: %s\n", err)
	os.Exit(1)
}

return sha1
}


func WriteTree(dirPath string) ([]TreeEntry, error) {
	var entries []TreeEntry

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		filePath := filepath.Join(dirPath, file.Name())
		if file.IsDir() {
			if file.Name() == ".myvcs" {
				continue
			}

			// RECURSIVELY call WriteTree for the subdirectory
			subdirEntries, err := WriteTree(filePath)
			if err != nil {
				return nil, err // Propagate error
			}

			// Create the tree object for the subdirectory
			subdirTreeSHA := CreateTree(subdirEntries)

			// Add an entry for the subdirectory to the current directory's entries
			entries = append(entries, TreeEntry{mode: "40000", name: file.Name(), sha1: subdirTreeSHA})

		} else {
			// File handling remains mostly the same, needs HashObject
			sha1 := HashObject(filePath) // Assuming HashObject is correct
			entries = append(entries, TreeEntry{mode: "100644", name: file.Name(), sha1: sha1})
		}
	}

	return entries, nil
}