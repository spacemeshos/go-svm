//go:build mage
// +build mage

// See <https://magefile.org/magefiles/> for more information about writing "magefiles".

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/magefile/mage/mg"
	artifacts "github.com/spacemeshos/go-svm/pkg/artifacts"
)

func libPath(artifactsDir string) string {
	path := filepath.Join(artifactsDir, "bins-Linux-release")
	path += ":" + filepath.Join(artifactsDir, "bins-macOS-release")
	path += ":" + filepath.Join(artifactsDir, "bins-Windows-release")
	return path
}

func DownloadArtifactsToDir(dir string) error {
	linuxZip := filepath.Join(dir, "bins-Linux-release.zip")
	gitkeep := filepath.Join(dir, ".gitkeep")

	if _, err := os.Stat(linuxZip); err == nil {
		fmt.Printf("Artifact files are already present.\n")
		return nil
	}

	os.RemoveAll(dir)
	os.MkdirAll(dir, os.ModePerm)
	os.OpenFile(gitkeep, os.O_RDONLY|os.O_CREATE, os.ModePerm)

	token := os.Getenv("GITHUB_TOKEN")
	fmt.Printf("Using the GitHub token '%s'", token)

	if err := artifacts.Download("master", token, dir); err != nil {
		log.Panic(err)
	}

	return nil
}

func Build() error {
	cmd := exec.Command("go", "mod", "download")
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func environWithLibPaths(here string) []string {
	artifacts := filepath.Join(here, "svm", "artifacts")
	libPath := libPath(artifacts)

	env := os.Environ()
	env = append(env, fmt.Sprintf("LD_LIBRARY_PATH=%s", libPath))
	env = append(env, fmt.Sprintf("DYLD_LIBRARY_PATH=%s", libPath))

	return env
}

func Install() error {
	here, _ := os.Getwd()
	dir := filepath.Join(here, "svm", "artifacts")

	mg.Deps(Build)
	mg.Deps(mg.F(DownloadArtifactsToDir, dir))

	cmd := exec.Command("go", "install", "-x", "./...")
	cmd.Env = environWithLibPaths(here)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Test() error {
	mg.Deps(Build)
	mg.Deps(Install)

	here, _ := os.Getwd()

	cmd := exec.Command("go", "test", "-p", "1", ".")
	cmd.Dir = filepath.Join(here, "svm")
	cmd.Env = environWithLibPaths(here)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
