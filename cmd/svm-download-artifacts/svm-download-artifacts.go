package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/go-github/v39/github"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var (
	branch string
	token  string
	dest   string
)

func httpClient(ctx context.Context) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return oauth2.NewClient(ctx, ts)
}

func main() {
	app := &cli.App{
		Name:      "svm-download-artifacts",
		Usage:     "Download the most recent SVM artifact files",
		Authors:   []*cli.Author{{Name: "The SVM team"}},
		Copyright: "2021",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "token",
				Usage:       "The GitHub authentication token. You can use 'gh auth status --show-token'.",
				Required:    true,
				Destination: &token,
			},
			&cli.StringFlag{
				Name:        "branch",
				Value:       "master",
				Usage:       "The repository branch from which to download artifact files.",
				Destination: &branch,
			},
			&cli.StringFlag{
				Name:        "dest",
				Usage:       "The destination directory.",
				Required:    true,
				Destination: &dest,
			},
		},
		Action: func(c *cli.Context) error {
			return mainAction(c)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func mainAction(c *cli.Context) error {
	ctx := context.Background()
	gh := github.NewClient(httpClient(ctx))

	workflow, _, err := gh.Actions.GetWorkflowByFileName(ctx, "spacemeshos", "svm", "ci.yml")
	if err != nil {
		return err
	}

	fmt.Printf("Workflow name: %s\n", *workflow.Name)

	runs, _, err := gh.Actions.ListWorkflowRunsByID(ctx, "spacemeshos", "svm", *workflow.ID, &github.ListWorkflowRunsOptions{Branch: branch, Status: "success"})
	if err != nil {
		return err
	}

	runID := *runs.WorkflowRuns[0].ID
	fmt.Printf("Workflow run ID: %d\n", runID)
	artifacts, _, err := gh.Actions.ListWorkflowRunArtifacts(ctx, "spacemeshos", "svm", runID, nil)
	if err != nil {
		return err
	}

	fmt.Printf("N. of artifacts: %d\n", *artifacts.TotalCount)
	os.RemoveAll(dest)
	os.MkdirAll(dest, 0777)
	for _, artifact := range artifacts.Artifacts {
		downloadArtifact(artifact, dest)
	}

	return nil
}

func downloadArtifact(artifact *github.Artifact, dir string) error {
	url := *artifact.ArchiveDownloadURL
	name := *artifact.Name
	path := filepath.Join(dir, name+".zip")

	fmt.Printf("Now downloading %s...\n", name)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	client := httpClient(context.Background())
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	nbytes, _ := io.Copy(out, resp.Body)
	fmt.Printf("Downloaded %d bytes. Now unzipping.\n", nbytes)

	return exec.Command("unzip", "-d", filepath.Join(dir, name), path).Run()
}
