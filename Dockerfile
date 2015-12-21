# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/starwars-countdown

# Set the public directory within the container end users can then docker mount
# images, or rebuild the image with their own images
ENV SWCD_PUBLIC_DIR=/go/src/starwars-countdown/public
ENV SWCD_BIND_ADDR=0.0.0.0:8080

# Build the server inside the container.
RUN go install starwars-countdown

# Run the server
ENTRYPOINT /go/bin/starwars-countdown

# The service listens on port 8080.
EXPOSE 8080

