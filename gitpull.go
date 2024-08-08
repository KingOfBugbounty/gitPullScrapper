package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	repoOwner = "projectdiscovery"
	repoName  = "nuclei-templates"
	saveDir   = "nuclei_templates"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable is not set")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	opts := &github.PullRequestListOptions{
		State: "open",
	}

	prs, _, err := client.PullRequests.List(ctx, repoOwner, repoName, opts)
	if err != nil {
		log.Fatalf("Error listing pull requests: %v", err)
	}

	fmt.Printf("Found %d open pull requests\n", len(prs))
	if len(prs) == 0 {
		return
	}

	os.MkdirAll(saveDir, os.ModePerm)

	for _, pr := range prs {
		files, _, err := client.PullRequests.ListFiles(ctx, repoOwner, repoName, *pr.Number, nil)
		if err != nil {
			log.Printf("Error listing files for PR #%d: %v", *pr.Number, err)
			continue
		}

		for _, file := range files {
			if strings.HasSuffix(*file.Filename, ".yaml") {
				err := downloadFile(*file.RawURL, filepath.Join(saveDir, filepath.Base(*file.Filename)))
				if err != nil {
					log.Printf("Error downloading file %s: %v", *file.Filename, err)
				} else {
					fmt.Printf("Downloaded file: %s\n", *file.Filename)
				}
			}
		}
	}
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath, data, 0644)
}
