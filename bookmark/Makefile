all: docker clean

bin:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bookmark server/*.go

docker: bin
	docker build -t "bookmark:`git rev-list --count master`-`git rev-parse HEAD | cut -c1-4`" -f Dockerfile .
	rm bookmark 2> /dev/null

deploy: docker
	docker save bookmark | bzip2 | pv | ssh dark 'bunzip2 | docker load'

compose: docker
	docker-compose up

clean:
	rm bookmark 2> /dev/null

.PHONY: bin docker all clean
