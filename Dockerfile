FROM golang:1.15.0 as builder

ENV CADDY_VERSION="v2.3.0"
WORKDIR /
RUN go get -u github.com/caddyserver/xcaddy/cmd/xcaddy
RUN echo "1" \
 && xcaddy build ${CADDY_VERSION} --with github.com/vantt/caddy-prometheus \
 && chmod +x /caddy 

ENV CADDY_VERSION="2.3.0"
FROM caddy:2.3.0-alpine
COPY --from=builder /caddy /usr/bin/caddy