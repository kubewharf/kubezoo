FROM --platform=${BUILDPLATFORM} golang:1.18 as builder
ADD . /build
ARG TARGETOS TARGETARCH GIT_VERSION
WORKDIR /build/
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GIT_VERSION=${GIT_VERSION} make build

FROM --platform=${TARGETPLATFORM} kubezoo/distroless-base-debian11:v1
ARG TARGETOS TARGETARCH
COPY --from=builder /build/_output/local/bin/${TARGETOS}/${TARGETARCH}/clusterresourcequota /usr/local/bin/clusterresourcequota
ENTRYPOINT ["/usr/local/bin/clusterresourcequota"]
