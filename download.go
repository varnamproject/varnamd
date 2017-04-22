package main

// Implements a downloader which dowloads words from the upstream page by page

type downloader struct {
	offset   int
	upstream UpstreamClient
	langCode string
	words    chan []*word
	err      chan error
}

func newDownloader(langCode string, offset int) *downloader {
	return &downloader{langCode: langCode, offset: offset, upstream: newUpstreamClient(langCode, "test"), words: make(chan []*word)}
}

func (d *downloader) start() error {
	_, err := d.upstream.GetCorpusDetails()
	if err != nil {
		return err
	}

	return nil
}

func (d *downloader) downloadWords() {

	//for {
	//offset := getDownloadOffset(langCode)
	//log.Printf("Offset: %d\n", offset)
	//if offset >= corpusSize {
	//break
	//}
	//filePath, err := downloadWordsAndUpdateOffset(langCode, offset)
	//if err != nil {
	//break
	//}
	//output <- filePath
	//}
	//log.Println("Local copy is upto date. No need to download from upstream")
	//close(output)
}
