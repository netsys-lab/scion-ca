FROM golang:1.18-alpine
WORKDIR /src
RUN apk add --no-cache git
RUN git clone https://github.com/smallstep/certificates.git
WORKDIR /src/certificates/cmd/step-ca

RUN CGO_ENABLED=0 go build

WORKDIR /src
RUN git clone https://github.com/smallstep/cli
WORKDIR /src/cli/cmd/step
RUN CGO_ENABLED=0 go build

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=0 /src/certificates/cmd/step-ca/step-ca /bin/step-ca
COPY --from=0 /src/cli/cmd/step/step /bin/step
ENTRYPOINT ["step-ca"]