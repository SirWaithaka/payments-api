FROM migrate/migrate

RUN apk add make

WORKDIR /app

COPY migrations migrations
COPY Makefile .

ENTRYPOINT ["make", "-s"]
CMD ["migrate"]
