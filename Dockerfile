FROM docker.io/golang:alpine3.15 AS build
WORKDIR /akc
ADD . /akc
RUN go mod tidy && env CGO_ENABLED=0 go build -trimpath -buildmode=pie -ldflags "-s -w" -o appoptics_kubernetes_controller

FROM alpine:latest
COPY --from=build --chown=1000:1000 /akc/appoptics_kubernetes_controller /appoptics_kubernetes_controller
USER 1000:1000
ENTRYPOINT ["/appoptics_kubernetes_controller"]