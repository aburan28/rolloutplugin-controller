package utils

import "net/http"

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
