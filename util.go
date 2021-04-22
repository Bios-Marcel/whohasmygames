package main

import (
	"io/ioutil"
	"net/http"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func responseToString(response *http.Response) string {
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	return string(bodyBytes)
}

func makeGetRequest(url string) (*http.Request, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("user-agent", "github.com/Bios-Marcel/whohasmygames@v1")

	return request, err
}
