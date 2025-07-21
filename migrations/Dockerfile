FROM migrate/migrate

RUN apk add make

WORKDIR /app

COPY migrations/sql migrations
COPY Makefile .

ENTRYPOINT ["make", "-s"]
CMD ["migrate"]
