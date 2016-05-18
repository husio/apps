all: docker clean

bin:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o smallpaste smallpaste.go

docker: bin
	docker build -t "smallpaste:`git rev-list --count master`-`git rev-parse HEAD | cut -c1-4`" -f Dockerfile .
	rm smallpaste 2> /dev/null

deploy: docker
	docker save smallpaste | bzip2 | pv | ssh dark 'bunzip2 | docker load'

clean:
	rm smallpaste 2> /dev/null

.PHONY: bin docker all clean
