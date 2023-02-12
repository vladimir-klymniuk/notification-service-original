FROM golang:1.14.4-stretch AS build

ADD . /src/
WORKDIR /src/
ARG private_key

ARG ENV=shmambamboo

# set ssh for stash
RUN if [ "$ENV" = "shmambamboo" ] ; then \
		git config --global url."git@someserver.com:".insteadOf "https://bamboo-stash.someserver.com/scm/" && \
		mkdir -p /root/.ssh && \
		echo "StrictHostKeyChecking no\n\nHost someserver.com\nHostName someserver.com\nPort 7999\nUser aveAdbroker" >> /root/.ssh/config && \
		echo "$private_key" > /root/.ssh/id_rsa && \
		chmod 600 /root/.ssh/id_rsa \
	; fi

RUN if [ "$ENV" = "shmambamboo" ] ; then \
		go get -u golang.org/x/lint/golint && \
		go get github.com/securego/gosec/cmd/gosec && \
		go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.26.0 \
	; fi

RUN if [ "$ENV" = "shmambamboo" ] ;\
    then \
		go test -v ./... && \
		golint ./... && \
		go vet ./...  && \
		go test ./... -coverprofile=cover.out && \
		go tool cover -func=cover.out && \
		gosec -exclude=G104,G301,G304 ./... \
	;fi


# CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GOPROXY=direct GOSUMDB=off go build -o /bin/notification-service -a -ldflags "-w -linkmode external -extldflags \"-static\" -s -X main.gitversion=$(git rev-parse HEAD) -X main.buildtime=$(date +%FT%T%z)" -installsuffix cgo main.go \
ENV GO111MODULE=on
RUN if [ "$ENV" = "shmambamboo" ] ; then \ 
	 	CGO_ENABLED=1 GOOS=linux GOPROXY=direct GOSUMDB=off go build -o /bin/notification-service-original -a -ldflags "-w -s -X main.gitversion=$(git rev-parse HEAD) -X main.buildtime=$(date +%FT%T%z)"  -installsuffix cgo main.go \
	; else \
		mv ./notification-service-original /bin/notification-service-original \
	; fi 

FROM golang:1.14.4-stretch

COPY --from=build /bin/notification-service /bin/notification-service
COPY ./docker-entrypoint.sh /

ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 11000
