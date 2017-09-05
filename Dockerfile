FROM golang:1.7.3
WORKDIR /go/src/github.com/IBM/ubiquity/
COPY . .
RUN go get -v github.com/Masterminds/glide
RUN glide up
RUN CGO_ENABLED=1 GOOS=linux go build -tags netgo -v -a --ldflags '-w -linkmode external -extldflags "-static"' -installsuffix cgo -o ubiquity main.go


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
RUN mkdir -p /tmp/ubiquity
COPY --from=0 /go/src/github.com/IBM/ubiquity/ubiquity .
COPY --from=0 /go/src/github.com/IBM/ubiquity/ubiquity-server.conf .
CMD ["./ubiquity"]
