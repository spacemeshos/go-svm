all: install build
.PHONY: all

ARTIFACTS_DIR := $(realpath svm)/artifacts
SVM_VERSION := 0.0.11

export GOOS
export CGO_CFLAGS = -I${ARTIFACTS_DIR}

ifeq ($(GOOS),windows)
	PLATFORM := windows
	SVM_CLI := svm-cli.exe
	export CGO_LDFLAGS = -L$(ARTIFACTS_DIR) -lsvm
	export PATH = $(PATH):$(ARTIFACTS_DIR)
else
	SVM_CLI := svm-cli
	ifeq ($(GOOS),darwin)
    	PLATFORM := macos
    	export CGO_LDFLAGS = $(ARTIFACTS_DIR)/libsvm.a -lm -ldl -framework Security -framework Foundation
	else
    	PLATFORM := linux
    	export CGO_LDFLAGS = $(ARTIFACTS_DIR)/libsvm.a -lm -ldl -Wl,-rpath,$(ARTIFACTS_DIR)
	endif
endif

svm/artifacts/svm-$(PLATFORM).zip:
	mkdir -p svm/artifacts/
	curl -L https://github.com/spacemeshos/svm/releases/download/v$(SVM_VERSION)/svm-$(PLATFORM)-v$(SVM_VERSION).zip -o svm/artifacts/svm-$(PLATFORM).zip
	unzip svm/artifacts/svm-$(PLATFORM).zip -d svm/artifacts/
	chmod +x svm/artifacts/$(SVM_CLI)
	ls svm/artifacts

clean:
	rm -rf svm/artifacts/
.PHONY: clean

download: svm/artifacts/svm-$(PLATFORM).zip
.PHONY: download

build: download
	go mod download
.PHONY: build

install: build download
	go install ./...
.PHONY: install

test: build install
	cd svm/inputs && ./generate_txs.sh
	cd svm && RUST_BACKTRACE=1 go test -v -p 1 .
.PHONY: test
