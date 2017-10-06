FROM golang:1.9.0
WORKDIR /go/src/github.com/IBM/ubiquity/
COPY . .
RUN go get -v github.com/Masterminds/glide
RUN glide up
RUN CGO_ENABLED=1 GOOS=linux go build -tags netgo -v -a --ldflags '-w -linkmode external -extldflags "-static"' -installsuffix cgo -o ubiquity main.go


FROM alpine:latest
RUN apk --no-cache add ca-certificates openssh
WORKDIR /root/
COPY --from=0 /go/src/github.com/IBM/ubiquity/ubiquity .
CMD ["./ubiquity"]
