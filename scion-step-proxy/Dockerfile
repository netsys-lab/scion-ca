FROM golang:1.18-alpine
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build

WORKDIR /step
RUN git clone https://github.com/netsys-lab/smallstep-cli-scion.git
WORKDIR /step/cmd/step
RUN CGO_ENABLED=0 go build -buildvcs=false


FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=0 /src/scion-step-proxy /bin/scion-step-proxy
COPY --from=0 /step/cmd/step /bin/step
ENTRYPOINT ["scion-step-proxy"]