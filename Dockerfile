FROM heroiclabs/nakama-pluginbuilder:3.22.0 AS builder

ENV GO111MODULE on
ENV CGO_ENABLED 1

WORKDIR /backend

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN go build --trimpath --mod=readonly --buildmode=plugin -o ./backend.so ./cmd/plugin/

FROM heroiclabs/nakama:3.22.0

USER root

RUN id -u nakama >/dev/null 2>&1 || (groupadd -g 1000 nakama && useradd -u 1000 -g nakama -m -s /bin/bash nakama)

COPY --from=builder --chown=nakama:nakama /backend/backend.so /nakama/data/modules/
COPY --chown=nakama:nakama local.yml /nakama/data/
COPY --chown=nakama:nakama entrypoint.sh /nakama/data/

RUN chmod +x /nakama/data/entrypoint.sh

USER nakama

ENTRYPOINT ["/nakama/data/entrypoint.sh"]
