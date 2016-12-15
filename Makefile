.DEFAULT_GOAL := all

DOCKER_REPO=thrawn01

# GO
GOPATH := $(shell go env | grep GOPATH | sed 's/GOPATH="\(.*\)"/\1/')
GLIDE := $(GOPATH)/bin/glide
PATH := $(GOPATH)/bin:$(PATH)
export $(PATH)

all: starwars-countdown

starwars-countdown: main.go
	go build -o starwars-countdown main.go

$(GLIDE):
	go get -u github.com/Masterminds/glide

get-deps: $(GLIDE)
	$(GLIDE) install

clean:
	rm starwars-countdown

build:
	docker build -t ${DOCKER_REPO}/starwars-countdown:latest .

run:
	@echo "Running Image on port 1313"
	-docker rm starwars-countdown
	docker run -p 1313:80 --name starwars-countdown ${DOCKER_REPO}/starwars-countdown:latest

publish:
	docker push ${DOCKER_REPO}/starwars-countdown:latest

all: build
