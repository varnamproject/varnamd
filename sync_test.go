package main

import (
	"errors"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
	libvarnam "github.com/varnamproject/libvarnam-golang"
)

var (
	pageSize = 2
)

type fakeUpstreamClient struct {
	corpusDetailsHandlerFunc func() (*libvarnam.CorpusDetails, error)
}

func (c *fakeUpstreamClient) GetCorpusDetails() (*libvarnam.CorpusDetails, error) {
	if c.corpusDetailsHandlerFunc != nil {
		return c.corpusDetailsHandlerFunc()
	}

	return &libvarnam.CorpusDetails{WordsCount: 6}, nil
}

func (c *fakeUpstreamClient) DownloadWords(offset int) (*page, error) {
	var words []*word
	if offset == 0 {
		words = []*word{
			&word{ID: 1, Confidence: 10, Word: "മലയാളം"},
			&word{ID: 2, Confidence: 10, Word: "മലയാളം"},
		}
	} else if offset == 2 {
		words = []*word{
			&word{ID: 3, Confidence: 11, Word: "മലയാളം"},
			&word{ID: 4, Confidence: 13, Word: "മലയാളം"},
		}
	} else if offset == 4 {
		words = []*word{
			&word{ID: 5, Confidence: 14, Word: "മലയാളം"},
			&word{ID: 6, Confidence: 16, Word: "മലയാളം"},
		}
	} else {
		panic(fmt.Sprintf("invalid offset %d", offset))
	}

	return &page{offset: offset, words: words}, nil
}

func TestStartFailsWhenGetCorpusDetailsFails(t *testing.T) {
	corpusdetails := func() (*libvarnam.CorpusDetails, error) {
		return nil, errors.New("test error")
	}

	s := newCorpusSync("ml", 10, pageSize)
	s.upstream = &fakeUpstreamClient{corpusDetailsHandlerFunc: corpusdetails}
	err := s.start()
	assert.Equal(t, "error getting corpus details for 'ml'. test error", err.Error())
}

func TestNothingToSync(t *testing.T) {
	corpusdetails := func() (*libvarnam.CorpusDetails, error) {
		return &libvarnam.CorpusDetails{WordsCount: 10}, nil
	}

	s := newCorpusSync("ml", 10, pageSize)
	s.upstream = &fakeUpstreamClient{corpusDetailsHandlerFunc: corpusdetails}
	err := s.start()
	assert.Equal(t, errNothingToSync, err)
}

func TestStartsAndReturnsNil(t *testing.T) {
	s := newCorpusSync("ml", 0, pageSize)
	s.upstream = &fakeUpstreamClient{}
	err := s.start()
	assert.Nil(t, err)
}

func TestSync(t *testing.T) {
	// initializing libvarnam channels
	initLanguageChannels()

	s := newCorpusSync("ml", 0, pageSize)
	s.upstream = &fakeUpstreamClient{}
	_ = s.start()

	var events []*syncProgress
	var errors []error
	s.progressHandler = func(p *syncProgress, e error) {
		if e != nil {
			errors = append(errors, e)
			return
		}
		events = append(events, p)
	}

	// TODO: Improve this
	for len(events) != 3 {
		// just waiting for execution to finish
	}

	assert.Equal(t, 3, len(events))
	assert.Equal(t, 0, len(errors))
	for _, event := range events {
		assert.Equal(t, 2, event.status.TotalWords)
		assert.Equal(t, 0, event.status.Failed)
	}
}
