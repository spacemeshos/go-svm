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
	"strings"

	"github.com/magefile/mage/mg"
)

func setLdLibraryPath(artifactsDir string) {
	ldVar := os.Getenv("LD_LIBRARY_PATH")
	ldPaths := strings.Split(ldVar, ":")

	if ldVar != "" {
		ldVar += ":"
	}
	if ldVar == "" || !strings.Contains(ldPaths[len(ldPaths)-1], artifactsDir) {
		ldVar = ldVar + filepath.Join(artifactsDir, "bins-Linux-release")
		ldVar = ldVar + ":" + filepath.Join(artifactsDir, "bins-macOS-release")
		ldVar = ldVar + ":" + filepath.Join(artifactsDir, "bins-Windows-release")
		os.Setenv("LD_LIBRARY_PATH", ldVar)
	}
}

func DownloadArtifactsToDir(dir string) error {
	if _, err := os.Stat(filepath.Join(dir, "bins-Linux-release.zip")); err == nil {
		fmt.Printf("Artifact files are already present.\n")
		return nil
	}

	os.RemoveAll(dir)
	os.MkdirAll(dir, os.ModePerm)
	os.OpenFile(filepath.Join(dir, ".gitkeep"), os.O_RDONLY|os.O_CREATE, 0666)

	script := "cmd/svm-download-artifacts/svm-download-artifacts.go"
	token := os.Getenv("GITHUB_TOKEN")
	cmd := exec.Command("go", "run", script, "--token", token, "--dest", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Panic(err)
	}

	return nil
}

func Build() error {
	here, _ := os.Getwd()
	dir := filepath.Join(here, "svm", "artifacts")
	mg.Deps(mg.F(DownloadArtifactsToDir, dir))

	setLdLibraryPath(dir)

	cmd := exec.Command("go", "mod", "download")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
	here, _ := os.Getwd()
	cmd.Dir = filepath.Join(here, "svm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
