all: docker clean

auth_service:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth-service cmd/auth-service/auth_service.go

register_account:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o register-account cmd/register-account/register_account.go

docker: auth_service register_account
	docker build -t "auth-service:`git rev-list --count master`-`git rev-parse HEAD | cut -c1-4`" -f Dockerfile .
	rm auth-service 2> /dev/null
	rm register-account 2> /dev/null

deploy: docker
	docker save auth-service | bzip2 | pv | ssh dark 'bunzip2 | docker load'

compose: docker
	docker-compose up

clean:
	rm auth-service 2> /dev/null
	rm register-account 2> /dev/null

.PHONY: bin docker all clean
