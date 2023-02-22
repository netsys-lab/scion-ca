FROM golang:1.18-alpine
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=0 /src/scion-step-proxy /bin/scion-step-proxy
ENTRYPOINT ["scion-step-proxy"]