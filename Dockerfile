# --------
# Stage 1: Retrieve and compile cmdhook
# --------

FROM golang:latest as builder

WORKDIR /go/src/github.com/gentlemanautomaton/cmdhook

COPY . .

RUN go install -v github.com/gentlemanautomaton/cmdhook

# --------
# Stage 2: Produce an artifact with just the compiled binary
# --------

FROM scratch

COPY --from=builder /go/bin/cmdhook /cmdhook
