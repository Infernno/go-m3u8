package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

func CheckFileExist(path string) (bool, error) {
	filePath, err := filepath.Abs(path)

	if err != nil {
		return false, err
	}

	file, err := os.Open(filePath)

	if err == nil {
		file.Close()
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

func GetOrDownloadParallel(path string, url string, lock *sync.WaitGroup) {
	lock.Add(1)
	defer lock.Done()

	GetOrDownload(path, url)
}

func GetOrDownload(path string, url string) (string, error) {
	filePath, _ := filepath.Abs(path)
	tempPath, _ := filepath.Abs(path + ".temp")

	file, err := os.Open(filePath)

	if err == nil {
		file.Close()
		return filePath, nil
	}

	file, err = os.OpenFile(tempPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)

	if err != nil {
		file.Close()
		return "", wrapError(err, "Failed to create a file at %s", tempPath)
	}

	err = downloadFile(url, file)
	file.Close()

	if err != nil {
		return "", wrapError(err, "Download failed")
	}

	err = os.Rename(tempPath, filePath)

	if err != nil {
		return "", wrapError(err, "Failed to rename a file")
	}

	return filePath, nil
}

func downloadFile(url string, file *os.File) error {
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status code %d", resp.StatusCode)
	}

	_, err = io.Copy(file, resp.Body)

	return err
}

func wrapError(err error, message string, args ...any) error {
	return errors.Join(fmt.Errorf(message+"\n", args), err)
}

func checkPath(path string) (string, error) {
	return "", nil
}
