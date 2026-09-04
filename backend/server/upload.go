package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxUploadSize = 10 << 20 // 10 MB

func saveUploadedImage(r *http.Request, fieldName string, folder string) (string, error) {
	file, _, err := r.FormFile(fieldName)
	if err == http.ErrMissingFile {
		return "", nil
	}

	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	contentType := http.DetectContentType(buffer[:n])

	var extension string

	switch contentType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/gif":
		extension = ".gif"
	default:
		return "", fmt.Errorf("unsupported image type")
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(folder, 0755)
	if err != nil {
		return "", err
	}

	fileName, err := generateSessionID()
	if err != nil {
		return "", err
	}

	fileName = fileName + extension
	fullPath := filepath.Join(folder, fileName)

	destination, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(destination, file)
	if err != nil {
		destination.Close()
		os.Remove(fullPath)

		return "", err
	}

	err = destination.Close()
	if err != nil {
		os.Remove(fullPath)

		return "", err
	}

	return "/" + filepath.ToSlash(fullPath), nil
}

func removeUploadedImage(imagePath string) {
	if imagePath == "" {
		return
	}

	filePath := strings.TrimPrefix(imagePath, "/")
	filePath = filepath.Clean(filePath)
	slashPath := filepath.ToSlash(filePath)

	if !strings.HasPrefix(slashPath, "uploads/") {
		log.Printf("refusing to remove invalid upload path: %s", imagePath)
		return
	}

	err := os.Remove(filePath)

	if err != nil &&
		!os.IsNotExist(err) {
		log.Printf("could not remove uploaded file %s: %v", imagePath, err)
	}
}
