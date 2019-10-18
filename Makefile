PROJECTNAME=$(shell basename "$(PWD)")

GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

build-server:
	go build -o $(GOBIN)/$(PROJECTNAME)-server main.go || exit

run-server: build-server
	$(GOBIN)/$(PROJECTNAME)-server

build-client:
	go build -o $(GOBIN)/$(PROJECTNAME)-client ./client/client.go || exit

run-client: build-client
	$(GOBIN)/$(PROJECTNAME)-client