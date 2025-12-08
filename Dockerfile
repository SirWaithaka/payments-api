# Using multi-stage builds
FROM golang:1.24-alpine AS builder

LABEL maintainer="Sir Waithaka"

RUN apk add git make

# Set the current working directory inside the container
WORKDIR /app

# copy go.{mod,sum} files for use to fetch dependencies
# fetching go dependencies first allows the build tool to cache this part of the image
COPY go.mod go.sum ./
COPY Makefile .
RUN go mod download

# install go tools
RUN make install-tools

# Copy project source files
COPY . ./

# Build the application
RUN make build


# Start the second image
FROM alpine:3

# set the working director in the container
WORKDIR /app

# copy binary
COPY --from=builder --chown=nonroot:nonroot /app/bin/payments /usr/bin/payments

ENTRYPOINT ["payments"]
CMD ["serve"]
