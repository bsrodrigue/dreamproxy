package fs

import (
	"io"
	"log"
	"os"
	"path"
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

func ResolveFilePath(target_path string, root_fs string) (string, os.FileInfo, error) {
	var err error
	var file_path string
	ext := path.Ext(target_path)
	ext = strings.ToLower(ext)

	// Page URLs
	if ext == "" {
		file_path = path.Join(root_fs, target_path)

		// Is Root
		if target_path == "/" {
			file_path = path.Join(root_fs, "index.html")
		}

	} else { // Resource URLs
		file_path = path.Join(root_fs, target_path)
	}

	stat, err := os.Stat(file_path)

	return file_path, stat, err
}
