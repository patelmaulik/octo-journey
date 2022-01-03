FROM golang:1.16-alpine AS builder

#RUN apk update && apk add bash

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

ENV PORT=8080

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# COPY *.go ./

# Build the application
RUN go build -a -installsuffix cgo -o main .

# Build healthcheck
#RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o healthcheck ./healthcheck

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp -r /build/main .

# RUN cp -r /build/healthcheck .

# Build a small image
#FROM scratch
FROM golang:1.16-alpine AS runtime
RUN apk update && apk add bash

COPY --from=builder /dist/main /
# COPY --from=builder /dist/healthcheck /
RUN apk add ca-certificates

# health check probe
# HEALTHCHECK \
#     --interval=1s --timeout=1s --start-period=2s --retries=3 \
#     CMD ["/healthcheck","http://localhost:8080/debug/vars"] || exit 1


EXPOSE 8080

#USER nonroot:nonroot

# Command to run
ENTRYPOINT ["/main"]