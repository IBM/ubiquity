FROM golang:1.9.0
WORKDIR /go/src/github.com/IBM/ubiquity/
COPY . .
RUN go get -v github.com/Masterminds/glide
RUN glide up
RUN CGO_ENABLED=1 GOOS=linux go build -tags netgo -v -a --ldflags '-w -linkmode external -extldflags "-static"' -installsuffix cgo -o ubiquity main.go


FROM alpine:latest
RUN apk --no-cache add ca-certificates=20161130-r2 openssl=1.0.2k-r0 openssh
WORKDIR /root/
COPY --from=0 /go/src/github.com/IBM/ubiquity/ubiquity .

COPY docker-entrypoint.sh .
RUN chmod 755 docker-entrypoint.sh

VOLUME /var/lib/ubiquity/ssl

ENV PATH=/root:$PATH

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["ubiquity"]

