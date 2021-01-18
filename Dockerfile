FROM golang:1.15.0 as builder

ENV CADDY_VERSION="v2.3.0"
WORKDIR /
RUN go get -u github.com/caddyserver/xcaddy/cmd/xcaddy
COPY ./ /src
RUN ls -l /src \
    && xcaddy build ${CADDY_VERSION} --with github.com/vantt/caddy-prometheus=/src \
    && chmod +x /caddy

FROM caddy:2.3.0-alpine
COPY --from=builder /caddy /usr/bin/caddy