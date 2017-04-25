package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/varnamproject/libvarnam-golang"
)

// Implements an upstream API client
type UpstreamClient interface {
	GetCorpusDetails() (*libvarnam.CorpusDetails, error)
	DownloadWords(offset int) (*page, error)
}

// A single page of words downloaded
type page struct {
	offset int
	words  []*word
}

type upstreamClient struct {
	langCode string
	url      string
}

func newUpstreamClient(langCode, url string) UpstreamClient {
	return &upstreamClient{langCode: langCode, url: url}
}

func (c *upstreamClient) GetCorpusDetails() (*libvarnam.CorpusDetails, error) {
	url := fmt.Sprintf("%s/meta/%s", c.url, c.langCode)
	var m metaResponse
	err := getJSONResponse(url, &m)
	if err != nil {
		return nil, err
	}
	return m.Result, nil
}

func (c *upstreamClient) DownloadWords(offset int) (*page, error) {
	url := fmt.Sprintf("%s/download/%s/%d", c.url, c.langCode, offset)
	var response downloadResponse
	err := getJSONResponse(url, &response)
	if err != nil {
		return nil, err
	}

	return &page{offset: offset, words: response.Words}, nil
}

func getJSONResponse(url string, output interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	jsonDecoder := json.NewDecoder(resp.Body)
	err = jsonDecoder.Decode(output)
	if err != nil {
		return err
	}
	return nil
}
