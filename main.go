package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

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

	for sleepTime := time.Millisecond; sleepTime < 1024*time.Millisecond; sleepTime *= 2 {
		body, cookies, err := getResp(req)
		if err != nil {
			time.Sleep(sleepTime)
			continue
		}

		for _, cookie := range cookies {
			http.SetCookie(w, cookie)
		}

		w.Write(body)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}

func getResp(req *http.Request) (respBody []byte, respCookies []*http.Cookie, err error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	doc := soup.HTMLParse(string(body))
	scripts := doc.FindAll("script", "id", "SIGI_STATE")
	if len(scripts) == 0 {
		return nil, nil, fmt.Errorf("failed to find SIGI_STATE")
	}

	return []byte(scripts[0].Text()), resp.Cookies(), nil
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
