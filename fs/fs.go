package fs

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func LoadFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	file_bin, err := io.ReadAll(file)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	return file_bin, err
}

// Possible paths:
// /foo/bar/
// /foo/bar
// /foo/bar/file.png
// /foo/bar/file.css
func ResolveFilePath(target_path string, root_fs string) (string, os.FileInfo, error) {
	var err error
	var file_path string

	file_path = filepath.Join(root_fs, filepath.Clean(target_path))

	if !strings.HasPrefix(file_path, root_fs) { // Path Traversal
		return "", nil, errors.New("Path Traversal")
	}

	// Is Root
	if target_path == "/" {
		file_path = filepath.Join(root_fs, "index.html")
	}

	stat, err := os.Stat(file_path)

	return file_path, stat, err
}
