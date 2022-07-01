# Get latest ca-certificates
FROM alpine AS certs
RUN apk --update add ca-certificates

FROM golang:1.18.3-alpine AS base

# To fix go get and build with cgo
RUN apk add --no-cache --virtual .build-deps \
    bash \
    gcc \
    git \
    musl-dev

RUN mkdir build
WORKDIR /build

COPY go.mod .
COPY go.sum .

RUN go mod download

# Build the image
FROM base as builder
COPY . .
ARG VERSION=dev-docker
ARG DATE=
ARG COMMIT=

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-X 'main.version=${VERSION}' -X 'main.date=${DATE}' -X 'main.commit=${COMMIT}' -extldflags '-static'" \
    -o dctna-server ./cmd/dctna-server

# Collect certificates and binary
FROM gcr.io/distroless/base-debian11
EXPOSE 8086 8443
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# root user required as the volumes mount as root
# files in the volumes can only be accessed by the owner of the files
# which are in this case root
# TODO: find a way arround this.
WORKDIR /root
VOLUME [ "/root/.notary", "/root/.docker/trust", "/root/certs" ]
COPY certs/ /root/certs/
COPY .notary/config.json /root/.notary/config.json
COPY --from=builder /build/dctna-server /root/
CMD [ "./dctna-server" ]
