FROM golang:1.9.6-alpine
MAINTAINER Roman Atachiants "roman@misakai.com"

# Copy the directory into the container.
RUN mkdir -p /go/src/github.com/emitter-io/emitter/
WORKDIR /go/src/github.com/emitter-io/emitter/
ADD . /go/src/github.com/emitter-io/emitter/

# Download and install any required third party dependencies into the container.
RUN apk add --no-cache g++ \ 
&& go-wrapper install \
&& apk del g++

# Expose emitter ports
EXPOSE 4000
EXPOSE 8080
EXPOSE 8443

# Start the broker
CMD ["go-wrapper", "run"]