package utils

import (
	"fmt"
	"net/http"
	"os"
)

type FileDownloader interface {
	Get(url string, header http.Header) (resp *http.Response, err error)
}

// FileDownloaderImpl is the default/real implementation of the FileDownloader interface
type FileDownloaderImpl struct {
}

func (fd FileDownloaderImpl) Get(url string, header http.Header) (resp *http.Response, err error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header = header
	return http.DefaultClient.Do(request)
}

func CheckIfExists(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
	}
	return nil
}
