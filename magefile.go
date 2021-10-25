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
	"runtime"

	"github.com/magefile/mage/mg"
	artifacts "github.com/spacemeshos/go-svm/download_artifacts"
)

func DownloadArtifactsToDir(dir string) error {
	svmCodecWASM := filepath.Join(dir, "svm_codec.wasm")
	gitkeep := filepath.Join(dir, ".gitkeep")

	if _, err := os.Stat(svmCodecWASM); err == nil {
		fmt.Printf("Artifact files are already present.\n")
		return nil
	}

	os.RemoveAll(dir)
	os.MkdirAll(dir, os.ModePerm)
	os.OpenFile(gitkeep, os.O_RDONLY|os.O_CREATE, os.ModePerm)

	token := os.Getenv("GITHUB_TOKEN")
	fmt.Printf("Using the GitHub token '%s'\n", token)

	if err := artifacts.Download("dylib-name", token, dir); err != nil {
		log.Panic(err)
	}
	os.Remove(filepath.Join(dir, "svm.lib"))

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
	cmd.Env = environCGo(here)
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
	cmd.Env = environCGo(here)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func environCGo(here string) []string {
	artifactsDir := filepath.Join(here, "svm", "artifacts")
	goos := runtime.GOOS

	env := os.Environ()
	env = append(env, fmt.Sprintf("CGO_CFLAGS=-I%s ", artifactsDir))
	if goos == "darwin" {
		env = append(env, fmt.Sprintf("CGO_LDFLAGS=%s/libsvm.a -lm -ldl -framework Security -framework Foundation", artifactsDir))
	} else if goos == "windows" {
		env = append(env, fmt.Sprintf("CGO_LDFLAGS=-L%s -lsvm", artifactsDir))
	} else {
		env = append(env, fmt.Sprintf("CGO_LDFLAGS=%s/libsvm.a -lm -ldl -Wl,-rpath,%s", artifactsDir, artifactsDir))
	}

	return env
}
