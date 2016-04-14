package varnamdaemon

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/varnamproject/libvarnam-golang"
)

type varnamWorker struct {
	LanguageCode string
	Text         string
	ResultWriter io.Writer
	Done         chan struct{}
}

type varnamTransliterationResult struct {
	Error  []string `json:"error"`
	Input  string   `json:"input"`
	Result []string `json:"result"`
}

type varnamReverseTransliterationResult struct {
	Error  []string `json:"error"`
	Result string   `json:"result"`
}

func NewVarnamWorker(langCode string, word string, writer io.Writer) varnamWorker {
	return varnamWorker{langCode, word, writer, make(chan struct{})}
}

var transliterationChannelMap = make(map[string]chan *varnamWorker)
var reverseTransliterationChannelMap = make(map[string]chan *varnamWorker)
var learningChannelMap = make(map[string]chan *varnamWorker)

func supportedLanguages() []string {
	return []string{"ta", "hi", "kn"}
}

func Init(learnOnly bool, maxProcs int) {

	for _, language := range supportedLanguages() {
		if learnOnly {
			//TODO: Launch launch only daemon here
		} else {
			transliterationChannelMap[language] = make(chan *varnamWorker)
			reverseTransliterationChannelMap[language] = make(chan *varnamWorker)
			learningChannelMap[language] = make(chan *varnamWorker)
			go learningWorker(language, learningChannelMap[language])

			for i := 0; i < maxProcs; i++ {
				go transliterationWorker(language, transliterationChannelMap[language])
				go reverseTransliterationWorker(language, reverseTransliterationChannelMap[language])
			}
		}
	}
}

func (v *varnamWorker) Transliterate() {
	transliterationChannelMap[v.LanguageCode] <- v
}

func (v *varnamWorker) ReverseTransliterate() {
	reverseTransliterationChannelMap[v.LanguageCode] <- v
}

func (v *varnamWorker) Learn() {
	learningChannelMap[v.LanguageCode] <- v
}

func transliterationWorker(langCode string, jobs <-chan *varnamWorker) {
	for j := range jobs {
		var result varnamTransliterationResult
		if j.ResultWriter == nil {
			panic("Writer is null")
		}
		handle, error := libvarnam.Init(langCode)
		if error != nil {
			result = varnamTransliterationResult{[]string{error.Error()}, j.Text, []string{}}
		} else {
			outputs, varnamError := handle.Transliterate(j.Text)
			if varnamError != nil {
				result = varnamTransliterationResult{[]string{varnamError.Error()}, j.Text, []string{}}
			} else {
				result = varnamTransliterationResult{[]string{}, j.Text, outputs}
			}
		}
		js, err := json.Marshal(result)
		if err != nil {
			return
		}

		fmt.Fprintf(j.ResultWriter, "%s", js)
		close(j.Done)
	}
}

func reverseTransliterationWorker(langCode string, jobs <-chan *varnamWorker) {
	for j := range jobs {
		var result varnamReverseTransliterationResult
		if j.ResultWriter == nil {
			panic("Writer is null")
		}
		handle, error := libvarnam.Init(langCode)
		if error != nil {
			result = varnamReverseTransliterationResult{[]string{error.Error()}, ""}
		} else {
			output, varnamError := handle.ReverseTransliterate(j.Text)
			if varnamError != nil {
				result = varnamReverseTransliterationResult{[]string{varnamError.Error()}, ""}
			} else {
				result = varnamReverseTransliterationResult{[]string{}, output}
			}
		}
		js, err := json.Marshal(result)
		if err != nil {
			return
		}

		fmt.Fprintf(j.ResultWriter, "%s", js)
		close(j.Done)
	}
}

func learningWorker(langCode string, jobs <-chan *varnamWorker) {
	var mutex = &sync.Mutex{}
	for j := range jobs {
		handle, error := libvarnam.Init(langCode)
		if error != nil {
			panic("Varnam not initialized")
		}
		mutex.Lock()
		varnamError := handle.Learn(j.Text)
		mutex.Unlock()
		//TODO: Instead of the above 3 lines, the worker should push the word to learn to the persistent store
		if varnamError != nil {
			panic("Learning of " + j.Text + " failed with error: " + varnamError.Error())
		}
		if j.ResultWriter == nil {
			panic("Writer is null")
		}
		fmt.Fprintf(j.ResultWriter, "Done")
		close(j.Done)
	}
}
