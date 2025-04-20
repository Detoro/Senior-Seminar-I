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
    fileContent, err := os.ReadFile(filePath) // Renamed to fileContent for clarity
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading file %s: %s\n", filePath, err)
        os.Exit(1) // Or return "", err
    }

    // Get the exact size of the byte slice
    contentSize := len(fileContent)

    // Construct the header as a byte slice
    header := []byte(fmt.Sprintf("blob %d\x00", contentSize))

    // Concatenate header bytes and file content bytes
    fullContent := append(header, fileContent...)

    // Calculate SHA-1 hash over the full byte slice content
    sha := sha1.Sum(fullContent)
    hash := fmt.Sprintf("%x", sha) // This hash is correct for this content

    // Derive object path
    objectPath := filepath.Join(".myvcs", "objects", hash[:2], hash[2:])

    // Check if object already exists
    if _, err := os.Stat(objectPath); err == nil {
        return hash // Return the hash if the file exists
    } else if !os.IsNotExist(err) {
        // Handle actual error other than "does not exist"
         fmt.Fprintf(os.Stderr, "Error stating object file %s: %s\n", objectPath, err)
        os.Exit(1) // Or return "", err
    }


    // Compress the full byte slice content
    var buffer bytes.Buffer
    z := zlib.NewWriter(&buffer)
    _, writeErr := z.Write(fullContent) // Write the full byte slice content
    closeErr := z.Close() // Close the writer to flush compressed data

    if writeErr != nil {
         fmt.Fprintf(os.Stderr, "Error writing to zlib writer for %s: %s\n", filePath, writeErr)
        os.Exit(1) // Or return "", writeErr
    }
    if closeErr != nil {
         fmt.Fprintf(os.Stderr, "Error closing zlib writer for %s: %s\n", filePath, closeErr)
        os.Exit(1) // Or return "", closeErr
    }


    // Ensure directory exists with more conventional permissions
    dir := filepath.Dir(objectPath)
    if err := os.MkdirAll(dir, 0755); err != nil { // Use 0755 permissions
         fmt.Fprintf(os.Stderr, "Error creating object directory %s: %s\n", dir, err)
        os.Exit(1) // Or return "", err
    }

    // Create and write the compressed data to the object file
    if err := os.WriteFile(objectPath, buffer.Bytes(), 0644); err != nil { // Use 0644 for file permissions
        fmt.Fprintf(os.Stderr, "Error writing object file %s: %s\n", objectPath, err)
        os.Exit(1) // Or return "", err
    }


    return hash // Return the calculated hash
}