ARG VERSION
# Build the Go Binary.
FROM golang:$VERSION as build_stage

WORKDIR /home
# Copy the source code into the container.
COPY ./ .
# Build the app binary.
RUN apt-get install git -y
RUN GIT_VERSION=$(git rev-list -1 HEAD); \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64  \
    go build -o banking-api -ldflags "-X main.version=$GIT_VERSION" app/main.go

# Run the Go Binary in Alpine.
FROM alpine:3.19 as application

WORKDIR /
COPY --from=build_stage /home/banking-api .
COPY --from=build_stage /home/app.env.example ./app.env
ENTRYPOINT ["/banking-api"]
