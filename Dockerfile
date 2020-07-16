# Setup a multi-stage build
FROM golang:1.13 as builder

# Using go modules, so purposely avoid building within /go directory
WORKDIR /build

# cache modules download in next two layers
COPY go.mod go.sum /build/
RUN go mod download

# now grab the remainder of source code
COPY . /build

# Build
# ...gcflags described by IntelliJ's remote Go debug info and https://skaffold.dev/docs/workflows/debug/#go
RUN CGO_ENABLED=0 go build -gcflags "all=-N -l"

###############################
# Runtime stage
# skaffold Go debugging doesn't seem to work with alpine or scratch
FROM ubuntu
COPY --from=builder /build/kube-metrics-reporter /usr/bin
# identify as Go for skaffold debug, see https://skaffold.dev/docs/workflows/debug/#go
ENV GOTRACEBACK=all
ENTRYPOINT ["/usr/bin/kube-metrics-reporter"]