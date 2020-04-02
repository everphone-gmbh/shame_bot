package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var httpClient http.Client
var githubRepo = os.Getenv("GITHUB_REPOSITORY")
var githubToken = os.Getenv("GITHUB_TOKEN")

type PrEntity struct {
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	Number    int64     `json:"number"`
}

type PrComment struct {
	Body string `json:"body"`
}

func githubRequest(request *http.Request) (*http.Response, error) {
	request.Header.Set("Accept", "application/vnd.github.shadow-cat-preview+json")
	request.Header.Set("Authorization", fmt.Sprintf("token %s", githubToken))
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func main() {
	now := time.Now()
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient = http.Client{Transport: customTransport, Timeout: time.Minute}
	log.Printf("listing PRs for repo %s\n", githubRepo)
	request, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/pulls", githubRepo), nil)
	if err != nil {
		log.Fatal(err)
	}
	response, err := githubRequest(request)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	var prList []PrEntity
	if err = json.Unmarshal(body, &prList); err != nil {
		log.Fatal(err)
	}
	for _, pr := range prList {
		if pr.CreatedAt.Before(now.AddDate(0, 0, -7)) {
			log.Printf("found a shameful PR! %+v\n", pr)
			// todo: check if we already shamed it
			// shaming
			age := time.Since(pr.CreatedAt)
			comment := PrComment{Body: fmt.Sprintf("this PR is %d days old! \n\n ![shame](https://user-images.githubusercontent.com/8014230/78236317-17403d80-74da-11ea-944a-2752e27620a8.gif)", int(age.Hours()/24))}
			jsonComment, err := json.Marshal(comment)
			if err != nil {
				log.Fatal(err)
			}
			request, err := http.NewRequest("POST", fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/comments", githubRepo, pr.Number), bytes.NewReader(jsonComment))
			if err != nil {
				log.Fatal(err)
			}
			response, err = githubRequest(request)
			if err != nil {
				log.Fatal(err)
			}
			if response.StatusCode != 201 {
				log.Fatal("failed to comment on a shameful PR, check the provided GITHUB_TOKEN")
			}
		}
	}
}
