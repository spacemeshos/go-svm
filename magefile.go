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

	script := "cmd/svm-download-artifacts/svm-download-artifacts.go"
	cmd := exec.Command("go", "run", script, "--token", token, "--dest", dir)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
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

func Install() error {
	here, _ := os.Getwd()
	dir := filepath.Join(here, "svm", "artifacts")

	mg.Deps(Build)
	mg.Deps(mg.F(DownloadArtifactsToDir, dir))

	cmd := exec.Command("go", "install", "./...")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("LD_LIBRARY_PATH=%s", libPath(dir)))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Test() error {
	mg.Deps(Build)
	mg.Deps(Install)

	here, _ := os.Getwd()
	dir := filepath.Join(here, "svm", "artifacts")

	cmd := exec.Command("go", "test", "-p", "1", ".")
	cmd.Dir = filepath.Join(here, "svm")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("LD_LIBRARY_PATH=%s", libPath(dir)))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DYLD_LIBRARY_PATH=%s", libPath(dir)))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
