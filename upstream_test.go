package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCorpusDetails(t *testing.T) {
	c := newUpstreamClient("ml", "https://api.varnamproject.com")
	corpusDetails, err := c.GetCorpusDetails()
	assert.Nil(t, err, "err should be nothing")
	assert.True(t, corpusDetails.WordsCount > 0)
}

func TestDownloadWords(t *testing.T) {
	c := newUpstreamClient("ml", "https://api.varnamproject.com")
	p, err := c.DownloadWords(0)
	assert.Nil(t, err, "err should be nothing")
	assert.True(t, len(p.words) > 0)
}
