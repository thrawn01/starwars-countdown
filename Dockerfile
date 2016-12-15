# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.7

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/thrawn01/starwars-countdown

# Set the public directory within the container end users can then docker mount
# images, or rebuild the image with their own images
ENV SWCD_PUBLIC_DIR=/go/src/github.com/thrawn01/starwars-countdown/public
ENV SWCD_BIND_ADDR=0.0.0.0:80

WORKDIR /go/src/github.com/thrawn01/starwars-countdown

# Build the server inside the container.
RUN go get .
RUN go install github.com/thrawn01/starwars-countdown

# Run the server
ENTRYPOINT /go/bin/starwars-countdown

# The service listens on port 80.
EXPOSE 80

