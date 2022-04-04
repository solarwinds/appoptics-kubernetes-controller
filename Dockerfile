FROM alpine:3.8

RUN adduser -D appoptics
RUN apk --update add ca-certificates
USER appoptics

ADD bin/appoptics/appoptics-kubernetes-controller-linux_amd64 /usr/local/bin/appoptics_kubernetes_controller-linux_amd64

ENTRYPOINT ["/usr/local/bin/appoptics_kubernetes_controller-linux_amd64"]
ARG VCS_REF
LABEL org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/solarwinds/appoptics_kubernetes_controller" \
      org.label-schema.license="Apache-2.0"