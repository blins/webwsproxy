FROM golang:1.18-alpine AS build
#ENV GO111MODULE=on
#ENV GOPATH=/go
WORKDIR /application
COPY . /application
RUN go mod download && \
    go mod verify && \
	go build -ldflags '-s -w' -o webwsproxy

FROM alpine
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /application/webwsproxy /usr/bin/webwsproxy
ENTRYPOINT ["/usr/bin/webwsproxy"]
EXPOSE 8000/tcp 8080/tcp
CMD [""]