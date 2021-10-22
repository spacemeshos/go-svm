package artifacts

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

func Download(branch, token, dest string) error {
	ctx := context.Background()
	gh := github.NewClient(httpClient(ctx, token))

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
		name := strings.ToLower(*artifact.Name)
		goos := runtime.GOOS
		if name == "svm_codec.wasm" {
			downloadArtifact(artifact, dest, token)
		} else if name == "bins-linux-release" && goos == "linux" {
			downloadArtifact(artifact, dest, token)
		} else if name == "bins-macos-release" && goos == "darwin" {
			downloadArtifact(artifact, dest, token)
		} else if name == "bins-windows-release" && goos == "windows" {
			downloadArtifact(artifact, dest, token)
		}
	}

	err = os.Chmod(filepath.Join(dest, "svm-cli"), 0755)
	if err != nil {
		return err
	}

	return nil
}

func downloadArtifact(artifact *github.Artifact, dir string, token string) error {
	url := *artifact.ArchiveDownloadURL
	name := *artifact.Name
	path := filepath.Join(dir, name+".zip")

	fmt.Printf("Now downloading %s...\n", name)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	client := httpClient(context.Background(), token)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	nbytes, _ := io.Copy(out, resp.Body)
	fmt.Printf("Downloaded %d bytes. Now unzipping.\n", nbytes)

	return exec.Command("unzip", "-d", dir, path).Run()
}

func httpClient(ctx context.Context, token string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return oauth2.NewClient(ctx, ts)
}
