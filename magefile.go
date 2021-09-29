//go:build mage
// +build mage

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kardianos/osext"
	"github.com/magefile/mage/mg"
)

func DownloadArtifactsToDir(dir string) error {
	if _, err := os.Stat(filepath.Join(dir, "bins-Linux-release.zip")); err == nil {
		// Artifact files are already present, let's return early.
		return nil
	}

	os.RemoveAll(dir)
	os.MkdirAll(dir, os.ModePerm)
	os.OpenFile(filepath.Join(dir, ".gitkeep"), os.O_RDONLY|os.O_CREATE, 0666)

	script := "cmd/svm-download-artifacts/svm-download-artifacts.go"
	token := os.Getenv("GITHUB_TOKEN")
	cmd := exec.Command("go", "run", script, "--token", token, "--dest", dir)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Panic(err)
	}

	return nil
}

func Build() error {
	here, _ := osext.ExecutableFolder()
	dir := filepath.Join(here, "svm", "artifacts")
	mg.Deps(mg.F(DownloadArtifactsToDir, dir))

	ldVar := os.Getenv("LD_LIBRARY_PATH")
	ldPaths := strings.Split(ldVar, ":")

	if ldVar == "" || !strings.Contains(ldPaths[len(ldPaths)-1], dir) {
		ldVar = ldVar + ":" + filepath.Join(dir, "bins-Linux-release")
		ldVar = ldVar + ":" + filepath.Join(dir, "bins-macOS-release")
		ldVar = ldVar + ":" + filepath.Join(dir, "bins-Windows-release")
		os.Setenv("LD_LIBRARY_PATH", ldVar)
	}

	cmd := exec.Command("go", "mod", "download")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func Install() error {
	mg.Deps(Build)

	cmd := exec.Command("go", "install", "./...")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func Test() error {
	mg.Deps(Build)
	mg.Deps(Install)

	fmt.Printf("LD_LIBRARY_PATH IS %s\n\n", os.Getenv("LD_LIBRARY_PATH"))

	cmd := exec.Command("go", "test", "-p", "1", ".")
	here, _ := osext.ExecutableFolder()
	cmd.Dir = filepath.Join(here, "svm")
	cmd.Stdout = os.Stdout
	err := cmd.Run()

	return err
}
