FROM golang:latest
ADD . /go/src/github.com/flant/promicher
WORKDIR /go/src/github.com/flant/promicher
RUN ./go-get.sh
RUN CGO_ENABLED=0 GOOS=linux ./go-install.sh

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=0 /go/bin/promicher .

# http://localhost:8000/api/v1/alerts is the default --destination-url
# 0.0.0.0:80 is the default --listen
CMD ["/promicher", "--labels=\".*\"", "--annotations=\".*\""]

EXPOSE 80
