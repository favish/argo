VERSION=`git describe --tags --abbrev=0`
BUILD=`date +%FT%T%z`

LDFLAGS=-ldflags "-X github.com/favish/argo/cmd.Version=${VERSION} -X github.com/favish/argo/cmd.Build=${BUILD}"

build:
	go build ${LDFLAGS} -o argo

clean:
	if [ -f argo ] ; then rm argo; fi