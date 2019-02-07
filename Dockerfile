FROM golang:1.10.1
RUN go get -d -v github.com/gree-gorey/io-exporter
WORKDIR /go/src/github.com/gree-gorey/io-exporter
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o io-exporter .

FROM alpine:3.7
WORKDIR /root/
COPY --from=0 /go/src/github.com/gree-gorey/io-exporter/io-exporter .
CMD ["./io-exporter"]
