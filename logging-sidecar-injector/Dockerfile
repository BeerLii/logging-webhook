# Build the webhook binary
FROM golang:1.19 as builder

RUN apt-get -y update

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY cmd/webhook/webhook.go webhook.go
COPY pkg/ pkg/


# Build
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.cn"

RUN go mod download
RUN go build -a -o webhook webhook.go


FROM alpine:3.9.2 as manager
COPY --from=builder /workspace/webhook .
ENTRYPOINT ["/webhook"]

