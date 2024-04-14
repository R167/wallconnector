FROM golang:1.22-alpine3.19
WORKDIR /src
COPY go.mod go.sum *.go /src/
COPY cmd /src/cmd

RUN go build -o /bin/prom ./cmd/prom

FROM alpine:3.19
COPY --from=0 /bin/prom /bin/prom

CMD exec /bin/prom -addr :80 -target $TARGET
