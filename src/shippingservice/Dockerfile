FROM golang:1.13-alpine as builder
RUN apk add --no-cache ca-certificates git && \
      wget -qO/go/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 && \
      chmod +x /go/bin/dep

ENV PROJECT github.com/GoogleCloudPlatform/microservices-demo/src/shippingservice
WORKDIR /go/src/$PROJECT

COPY . .
RUN go install .

FROM alpine as release
RUN apk add --no-cache ca-certificates
RUN GRPC_HEALTH_PROBE_VERSION=v0.2.0 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe
COPY --from=builder /go/bin/shippingservice /shippingservice
ENV PORT=50051
ENV HEALTH_PORT=50053
EXPOSE 50051 50052 50053
ENTRYPOINT ["/shippingservice"]
