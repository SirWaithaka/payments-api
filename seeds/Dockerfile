FROM alpine/psql

RUN apk add make

WORKDIR /app

COPY seeds ./seeds
COPY Makefile .

ENTRYPOINT ["make", "-s"]
CMD ["seed"]
