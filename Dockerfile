FROM golang:1.11.3-alpine AS builder

RUN apk add --no-cache ca-certificates git curl

RUN curl -fsSL -o /usr/local/bin/dep \
    https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 && \
    chmod +x /usr/local/bin/dep

COPY . /go/src/gitea.fge.cloud/fabian_gehrlicher/reverseproxy
WORKDIR /go/src/gitea.fge.cloud/fabian_gehrlicher/reverseproxy

RUN dep ensure -vendor-only
RUN CGO_ENABLED=0 go build \
    -installsuffix 'static' \
    -o /reverseproxy .
RUN go build .

FROM nginx:1.15.2 as final

EXPOSE 5000
COPY --from=builder /reverseproxy /reverseproxy
RUN chmod +x /reverseproxy
USER root

ENTRYPOINT ["/reverseproxy"]
