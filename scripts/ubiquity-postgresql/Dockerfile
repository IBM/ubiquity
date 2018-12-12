FROM postgres:9.6

ENV UBIQUITY_DB_CERT_PRIVATE="/var/lib/postgresql/ssl/private/ubiquity-db.key" \
    UBIQUITY_DB_CERT_PUBLIC="/var/lib/postgresql/ssl/private/ubiquity-db.crt" \
    UBIQUITY_DB_PROVIDED_CERT_PRIVATE="/var/lib/postgresql/ssl/provided/ubiquity-db.key" \
    UBIQUITY_DB_PROVIDED_CERT_PUBLIC="/var/lib/postgresql/ssl/provided/ubiquity-db.crt"\
    UBIQUITY_DB_MAX_CONNECTION="1000"

RUN PGSSL_PRIVATE="`dirname $UBIQUITY_DB_CERT_PRIVATE`" && mkdir -p "$PGSSL_PRIVATE" && chown -R postgres:postgres "$PGSSL_PRIVATE" && chmod 777 "$PGSSL_PRIVATE"
RUN PGSSL_PUBLIC="`dirname $UBIQUITY_DB_CERT_PUBLIC`" && mkdir -p "$PGSSL_PUBLIC" && chown -R postgres:postgres "$PGSSL_PUBLIC" && chmod 777 "$PGSSL_PUBLIC"

ADD docker-entrypoint-initdb.d /docker-entrypoint-initdb.d/

COPY ubiquity-docker-entrypoint.sh /usr/local/bin/
COPY LICENSE .
COPY notices_file_for_ibm_storage_enabler_for_containers_db ./NOTICES
RUN ln -s usr/local/bin/ubiquity-docker-entrypoint.sh / # backwards compat

ENTRYPOINT ["ubiquity-docker-entrypoint.sh"]

EXPOSE 5432
CMD ["postgres"]

