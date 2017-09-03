FROM golang:1.9-alpine
MAINTAINER Roman Atachiants "roman@misakai.com"

# Copy the directory into the container.
RUN mkdir -p /go/src/emitter
WORKDIR /go/src/emitter
ADD . /go/src/emitter/

# Download and install any required third party dependencies into the container.
RUN go-wrapper install

# Expose emitter ports
EXPOSE 4000
EXPOSE 8080
EXPOSE 8443

# Start the broker
CMD ["go-wrapper", "run"]