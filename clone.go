package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func cloneRepo(zipURL string) error {
	fmt.Println("Cloning from:", zipURL)
	resp, err := http.Get(zipURL)
	if err != nil {
		return fmt.Errorf("failed to fetch repo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	r, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return fmt.Errorf("failed to read zip: %w", err)
	}

	for _, f := range r.File {
		fpath := filepath.Join(".myvcs", strings.TrimPrefix(f.Name, ".myvcs/"))
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("mkdir error: %w", err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("file create error: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("file open error: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return fmt.Errorf("copy error: %w", err)
		}
	}

	fmt.Println("Repository cloned into .myvcs/")
	return nil
}
