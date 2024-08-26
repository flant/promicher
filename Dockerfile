FROM golang:latest@sha256:613a108a4a4b1dfb6923305db791a19d088f77632317cfc3446825c54fb862cd
ADD . /go/src/github.com/flant/promicher
WORKDIR /go/src/github.com/flant/promicher
RUN ./go-get.sh
RUN CGO_ENABLED=0 GOOS=linux ./go-install.sh

FROM alpine:latest@sha256:0a4eaa0eecf5f8c050e5bba433f58c052be7587ee8af3e8b3910ef9ab5fbe9f5
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=0 /go/bin/promicher .

# http://localhost:8000/api/v1/alerts is the default --destination-url
# 0.0.0.0:80 is the default --listen
CMD ["/promicher", "--labels=\".*\"", "--annotations=\".*\""]

EXPOSE 80
