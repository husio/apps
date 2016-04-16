all: docker clean

bin:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o feedreaderd cmd/feedreaderd/feedreaderd.go

docker: bin
	docker build -t "feedreader:`git rev-list --count master`-`git rev-parse HEAD | cut -c1-4`" -f Dockerfile .
	rm feedreaderd 2> /dev/null

deploy: docker
	docker save feedreader | bzip2 | pv | ssh dark 'bunzip2 | docker load'

compose: docker
	docker-compose up

clean:
	rm feedreader 2> /dev/null

.PHONY: bin docker all clean
