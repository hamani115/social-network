package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	defer destination.Close()

	_, err = io.Copy(destination, file)
	if err != nil {
		return "", err
	}

	return "/" + filepath.ToSlash(fullPath), nil
}
