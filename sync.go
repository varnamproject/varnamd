package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	libvarnam "github.com/varnamproject/libvarnam-golang"
)

var errNothingToSync = errors.New("nothing to sync")

type progressFunc func(*syncProgress, error)

type corpusSync struct {
	langCode         string
	offset, pageSize int
	upstream         UpstreamClient
	progressHandler  progressFunc
}

type syncProgress struct {
	offset   int
	langCode string
	status   *libvarnam.LearnStatus
}

func newCorpusSync(langCode string, offset, pageSize int) *corpusSync {
	return &corpusSync{
		langCode: langCode,
		offset:   offset,
		pageSize: pageSize,
		upstream: defaultUpstreamClient(langCode),
	}
}

func (s *corpusSync) start() error {
	corpusDetails, err := s.upstream.GetCorpusDetails()
	if err != nil {
		return fmt.Errorf("error getting corpus details for '%s'. %v", s.langCode, err)
	}

	offset := s.offset

	if offset >= corpusDetails.WordsCount {
		return errNothingToSync
	}

	go func() {
		for p := range s.downloadAllWordsFrom(offset, corpusDetails.WordsCount) {
			status, err := s.learnWords(p)
			if s.progressHandler == nil {
				continue
			}
			if err != nil {
				s.progressHandler(nil, err)
				continue
			}
			s.progressHandler(&syncProgress{
				offset:   offset,
				langCode: s.langCode,
				status:   status,
			}, nil)
		}
	}()

	return nil
}

func (s *corpusSync) downloadAllWordsFrom(offset, maxOffset int) <-chan *page {
	const maxNoOfRetries = 10
	pages := make(chan *page)
	retries := maxNoOfRetries

	go func() {
		defer close(pages)
		for offset < maxOffset {
			p, err := s.upstream.DownloadWords(offset)
			if err != nil {
				retries = retries - 1
				time.Sleep(10 * time.Second)
				if retries == 0 {
					if s.progressHandler != nil {
						s.progressHandler(nil, fmt.Errorf("Failed to download words. offset: %d, error: %v", offset, err))
					}
					return
				}
				continue
			}

			retries = maxNoOfRetries
			offset = offset + len(p.words)
			pages <- p
		}
	}()

	return pages
}

func (s *corpusSync) learnWords(p *page) (*libvarnam.LearnStatus, error) {
	f, err := transformAndPersistWords(s.langCode, p)
	if err != nil {
		return nil, fmt.Errorf("Failed to persist words to file for learning. %v", err)
	}

	var ls *libvarnam.LearnStatus

	_, err = getOrCreateHandler(s.langCode, func(handle *libvarnam.Varnam) (data interface{}, err error) {
		learnStatus, err := handle.LearnFromFile(f)
		if err != nil {
			return nil, fmt.Errorf("Error learning from '%s'. %v", f, err)
		}

		ls = learnStatus
		os.Remove(f)
		return
	})

	return ls, err
}

func transformAndPersistWords(langCode string, p *page) (string, error) {
	tmpDir := os.TempDir()
	targetFile, err := os.Create(path.Join(tmpDir, fmt.Sprintf("%s.%d", langCode, p.offset)))
	if err != nil {
		return "", err
	}
	defer targetFile.Close()

	for _, word := range p.words {
		_, err = targetFile.WriteString(fmt.Sprintf("%s %d\n", word.Word, word.Confidence))
		if err != nil {
			return "", err
		}
	}
	return targetFile.Name(), nil
}
