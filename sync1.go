package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	libvarnam "github.com/varnamproject/libvarnam-golang"
)

var errNothingToSync = errors.New("nothing to sync")

type wordsSync struct {
	langCode string
	upstream UpstreamClient
}

func (s *wordsSync) start(offset int) error {
	corpusDetails, err := s.upstream.GetCorpusDetails()
	if err != nil {
		log.Printf("Error getting corpus details for '%s'. %s\n", s.langCode, err.Error())
		return err
	}

	if offset >= corpusDetails.WordsCount {
		return errNothingToSync
	}

	go func() {
		pages := s.downloadAllWords(offset)
		for {
			p, more := <-pages
			if !more {
				return
			}
			s.learnWords(p)
		}
	}()

	return nil
}

func (s *wordsSync) downloadAllWords(offset int) <-chan *page {
	const maxNoOfRetries = 10
	pages := make(chan *page)
	retries := maxNoOfRetries

	go func() {
		defer close(pages)
		for {
			p, err := s.upstream.DownloadWords(offset)
			if err != nil {
				log.Printf("[%s] Failed to download words. offset: %d, error: %s\n", s.langCode, offset, err)
				retries = retries - 1
				time.Sleep(10 * time.Second)
				if retries == 0 {
					return
				}
				continue
			}

			retries = maxNoOfRetries
			pages <- p
		}
	}()

	return pages
}

func (s *wordsSync) learnWords(p *page) {
	f, err := transformAndPersistWords(s.langCode, p)
	if err != nil {
		log.Printf("[%s] Failed to persist words to file for learning. %s", s.langCode, err)
		return
	}

	getOrCreateHandler(s.langCode, func(handle *libvarnam.Varnam) (data interface{}, err error) {
		learnStatus, err := handle.LearnFromFile(f)
		if err != nil {
			log.Printf("Error learning from '%s'\n", err.Error())
		} else {
			log.Printf("Learned from '%s'. TotalWords: %d, Failed: %d\n", f, learnStatus.TotalWords, learnStatus.Failed)
		}

		err = os.Remove(f)
		if err != nil {
			log.Printf("Error deleting '%s'. %s\n", f, err.Error())
		}

		return
	})
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
