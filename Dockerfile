FROM golang:alpine AS builder
MAINTAINER Roman Atachiants "roman@misakai.com"

# Copy the directory into the container.
RUN mkdir -p /go/src/github.com/emitter-io/emitter/
WORKDIR /go/src/github.com/emitter-io/emitter/
ADD . /go/src/github.com/emitter-io/emitter/

# Download and install any required third party dependencies into the container.
RUN apk add --no-cache g++ \
  && go install \
  && apk del g++

# Base image for runtime
FROM alpine:latest

WORKDIR /root/
# Get the executable binary from build-img declared previously
COPY --from=builder /go/bin/emitter .

# Expose emitter ports
EXPOSE 4000
EXPOSE 8080
EXPOSE 8443

# Start the broker
CMD ["./emitter"]
