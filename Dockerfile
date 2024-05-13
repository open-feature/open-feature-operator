# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.20.3-alpine3.16 AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

COPY apis/ apis/
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go work init ./apis && go mod download

# Copy the go source
COPY main.go main.go
COPY webhooks/ webhooks/
COPY controllers/ controllers/
COPY common/ common/

ARG TARGETOS
ARG TARGETARCH

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot as production
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
