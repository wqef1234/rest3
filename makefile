.PHONY: run
run:
	go run main.go

.PHONY:build
build:
	go build main.go

.PHONY:exec
exec:
	./main

.PHONY: docker_build
docker_build:
	sudo docker build -t rest2 .

.PHONY: docker_run
docker_run:
	sudo docker run -p 8081:8080 -it rest2

.PHONY: images
images:
	sudo docker images


DEFAULT_GOAL := run