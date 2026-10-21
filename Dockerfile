# --------
# Stage 1: Retrieve and compile cmdhook
# --------

FROM golang:1.27-alpine AS builder

RUN apk --no-cache add git

WORKDIR /go/src/github.com/gentlemanautomaton/cmdhook
COPY . .

# Disable CGO to make sure we don't rely on libc
ENV CGO_ENABLED=0

# Exclude debugging symbols and set the netgo tag for Go-based DNS resolution
ENV BUILD_FLAGS="-v -a -ldflags '-d -s -w' -tags netgo"

RUN go get -d -v ./...
RUN go install -v ./...

# --------
# Stage 2: Produce an artifact with just the compiled binary
# --------

FROM scratch

COPY --from=builder /go/bin/cmdhook /cmdhook
