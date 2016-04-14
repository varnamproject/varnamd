package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kumaranvram/varnamdaemon"
	flag "github.com/ogier/pflag"
)

var (
	learnOnly bool
	port      int
	maxProcs  int
)

func init() {
	flag.BoolVarP(&learnOnly, "learn-only", "l", false, "Starts a learn only daemon")
	flag.IntVarP(&port, "port", "p", 3000, "Starts the daemon in a specific port")
	flag.IntVarP(&maxProcs, "max-procs", "n", 1, "Number of go routines for each module. Useful only if not started with -learn-only switch")
}

func main() {
	flag.Parse()

	varnamdaemon.Init(learnOnly, maxProcs)
	startServer(learnOnly, port)
}

func startServer(learnOnly bool, port int) {
	r := mux.NewRouter()
	if learnOnly {
		//TODO: Launch LearnOnlyDaemon
	} else {
		fmt.Println("TL and RTL services started")
		r.HandleFunc("/tl/{langCode}/{word}", transliterationHandler).Methods("GET")
		r.HandleFunc("/rtl/{langCode}/{word}", reverseTransliterationHandler).Methods("GET")
		r.HandleFunc("/learn", learnHandler).Methods("POST")
	}
	http.Handle("/", r)
	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)

}

func learnHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	langCode := r.Form.Get("langCode")
	word := r.Form.Get("word")

	worker := varnamdaemon.NewVarnamWorker(langCode, word, w)
	worker.Learn()
	<-worker.Done
}

func transliterationHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	langCode := params["langCode"]
	word := params["word"]

	worker := varnamdaemon.NewVarnamWorker(langCode, word, w)
	worker.Transliterate()
	<-worker.Done
}

func reverseTransliterationHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	langCode := params["langCode"]
	word := params["word"]

	worker := varnamdaemon.NewVarnamWorker(langCode, word, w)
	worker.ReverseTransliterate()
	<-worker.Done
}
