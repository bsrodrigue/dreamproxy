package file_system

import (
	"io"
	"log"
	"os"
)

func LoadFile(filepath string) ([]byte, error) {
	index_file, err := os.Open(filepath)

	if err != nil {
		log.Println(err)
		return []byte{}, err
	}

	index_content, err := io.ReadAll(index_file)

	if err != nil {
		return []byte{}, err
	}

	defer index_file.Close()

	return index_content, err
}
