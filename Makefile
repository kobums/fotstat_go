tag=latest

all: server

server:
	go build -o bin/fotstat main.go

run:
	go run main.go

test:
	go test -v ./...

linux:
	env GOOS=linux GOARCH=amd64 go build -o bin/fotstat.linux main.go

dockerbuild:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s' -o bin/fotstat.linux main.go

docker: dockerbuild
	docker build --platform linux/amd64 -t kobums/fotstat:$(tag) .

dockerrun:
	docker run --env-file .env --platform linux/amd64 -d --name="fotstat" -p 8007:8007 kobums/fotstat:$(tag)

push: docker
	docker push kobums/fotstat:$(tag)

clean:
	rm -f bin/fotstat