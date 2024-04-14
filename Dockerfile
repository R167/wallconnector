FROM golang:1.22-alpine3.19
WORKDIR /src
COPY go.mod go.sum *.go /src/
COPY cmd /src/cmd

RUN go build -o /bin/prom ./cmd/prom

FROM alpine:3.19
COPY --from=0 /bin/prom /bin/prom

LABEL org.opencontainers.image.source=https://github.com/R167/wallconnector
LABEL org.opencontainers.image.description="Prometheus metrics for Tesla wall connector proxy."
LABEL org.opencontainers.image.licenses=MIT

CMD exec /bin/prom -addr :80 -target $TARGET
