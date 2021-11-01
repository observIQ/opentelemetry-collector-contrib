FROM alpine:latest as prep
RUN apk --update add ca-certificates

ARG USER_UID=10001

RUN mkdir -p /tmp /opt

FROM debian:stretch-slim

USER ${USER_UID}

RUN apt-get update && apt-get install -y ca-certificates openjdk-8-jre

COPY --from=prep /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY opentelemetry-java-contrib-jmx-metrics.jar /opt/opentelemetry-java-contrib-jmx-metrics.jar
COPY otelcontribcol /
EXPOSE 4317 55680 55679
ENTRYPOINT ["/otelcontribcol"]
CMD ["--config", "/etc/otel/config.yaml"]
