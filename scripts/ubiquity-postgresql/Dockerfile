FROM postgres:9.6

ENV PGSSL /var/lib/postgresql/ssl
RUN mkdir -p "$PGSSL" && chown -R postgres:postgres "$PGSSL" && chmod 777 "$PGSSL"

VOLUME /var/lib/postgresql/ssl

ADD docker-entrypoint-initdb.d /docker-entrypoint-initdb.d/

