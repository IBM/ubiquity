FROM golang:1.9.1
WORKDIR /go/src/github.com/IBM/ubiquity/
COPY . .
RUN go get -v github.com/Masterminds/glide
RUN glide up
RUN CGO_ENABLED=1 GOOS=linux go build -tags netgo -v -a --ldflags '-w -linkmode external -extldflags "-static"' -installsuffix cgo -o ubiquity main.go


FROM alpine:3.7
RUN apk --no-cache add ca-certificates=20171114-r0 openssl=1.0.2n
WORKDIR /root/
COPY --from=0 /go/src/github.com/IBM/ubiquity/ubiquity .
COPY --from=0 /go/src/github.com/IBM/ubiquity/LICENSE .
COPY --from=0 /go/src/github.com/IBM/ubiquity/scripts/notices_file_for_ibm_storage_enabler_for_containers ./NOTICES

COPY docker-entrypoint.sh .
RUN chmod 755 docker-entrypoint.sh

# comments below should be removed when we implement the new SSL_MODE env variable
ENV PATH=/root:$PATH \
    UBIQUITY_SERVER_CERT_PRIVATE=/var/lib/ubiquity/ssl/private/ubiquity.key \
    UBIQUITY_SERVER_CERT_PUBLIC=/var/lib/ubiquity/ssl/private/ubiquity.crt \
    UBIQUITY_SERVER_VERIFY_SCBE_CERT=/var/lib/ubiquity/ssl/public/scbe-trusted-ca.crt \
    UBIQUITY_DB_SSL_ROOT_CERT=/var/lib/ubiquity/ssl/public/ubiquity-db-trusted-ca.crt \
    SSL_MODE=verify-full

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["ubiquity"]

