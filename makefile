VERSION=`git describe --tags --abbrev=0`
BUILD=`date +%FT%T%z`

LDFLAGS=-ldflags "-X github.com/favish/argo/cmd.Version=${VERSION} -X github.com/favish/argo/cmd.Build=${BUILD}"

build:
	go build ${LDFLAGS} -o argo
	env GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o argo-nix

clean:
	if [ -f argo ] ; then rm argo; fi