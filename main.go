package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/anaskhan96/soup"
	"github.com/gorilla/mux"
)

func requestHandler(w http.ResponseWriter, r *http.Request) {
	postID := mux.Vars(r)["postID"]
	w.Header().Set("Content-Type", "application/json")

	// The author field doesn't actually matter
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.tiktok.com/@a/video/%s", postID), nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	doc := soup.HTMLParse(string(body))
	scripts := doc.FindAll("script", "id", "sigi-persisted-data")

	for _, script := range scripts {
		text := strings.TrimPrefix(script.Text(), "window['SIGI_STATE']=")

		idx := strings.Index(text, ";window['SIGI_RETRY']")
		text = text[:idx]

		w.Write([]byte(text))
	}
}

func main() {
	f, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)

	r := mux.NewRouter()
	r.HandleFunc("/tiktok/video/{postID}", requestHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
