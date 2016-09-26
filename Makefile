.DEFAULT_GOAL := test

.PHONY : build

ifeq ($(GOOS),windows)
DEST = build/credhub.exe
else
DEST = build/credhub
endif

GOFLAGS := -o $(DEST)

dependencies :
		go get github.com/onsi/ginkgo/ginkgo
		go get golang.org/x/tools/cmd/goimports
		go get github.com/maxbrunsfeld/counterfeiter
		go get -u github.com/kardianos/govendor
		govendor sync

format : dependencies
		goimports -w .
		go fmt .

ginkgo : dependencies
		ginkgo -r -randomizeSuites -randomizeAllSpecs -race 2>&1

test : format ginkgo

ci : ginkgo

build :
		mkdir -p build
		go build $(GOFLAGS)
