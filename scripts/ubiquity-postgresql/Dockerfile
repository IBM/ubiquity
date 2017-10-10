FROM postgres:9.6

ENV PGSSL /var/lib/postgresql/ssl
RUN mkdir -p "$PGSSL" && chown -R postgres:postgres "$PGSSL" && chmod 777 "$PGSSL"

VOLUME /var/lib/postgresql/ssl

ADD docker-entrypoint-initdb.d /docker-entrypoint-initdb.d/

COPY ubiquity-docker-entrypoint.sh /usr/local/bin/
RUN ln -s usr/local/bin/ubiquity-docker-entrypoint.sh / # backwards compat

ENTRYPOINT ["ubiquity-docker-entrypoint.sh"]

EXPOSE 5432
CMD ["postgres"]

