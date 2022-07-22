FROM --platform=${BUILDPLATFORM} golang:1.17 as builder
ADD . /build
ARG TARGETOS TARGETARCH GIT_VERSION
WORKDIR /build/
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GIT_VERSION=${GIT_VERSION} make build

FROM --platform=${TARGETPLATFORM} kubezoo/distroless-base-debian11
ARG TARGETOS TARGETARCH
COPY --from=builder /build/_output/local/bin/${TARGETOS}/${TARGETARCH}/kubezoo /usr/local/bin/kubezoo
ENTRYPOINT ["/usr/local/bin/kubezoo"]
