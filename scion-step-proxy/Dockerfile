FROM golang:1.18-alpine
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build
RUN apk add --no-cache git
WORKDIR /
RUN git clone https://github.com/netsys-lab/smallstep-cli-scion.git
RUN git clone https://github.com/netsys-lab/smallstep-ca-scion.git
WORKDIR /smallstep-cli-scion/cmd/step
RUN CGO_ENABLED=0 go build -buildvcs=false

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=0 /src/scion-step-proxy /bin/scion-step-proxy
COPY --from=0 /smallstep-cli-scion/cmd/step/step /bin/step
ENTRYPOINT ["scion-step-proxy"]