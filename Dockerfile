FROM heroiclabs/nakama-pluginbuilder:3.22.0 AS builder

ENV GO111MODULE on
ENV CGO_ENABLED 1

WORKDIR /backend

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN go build --trimpath --mod=readonly --buildmode=plugin -o ./backend.so ./cmd/plugin/

FROM heroiclabs/nakama:3.22.0

COPY --from=builder /backend/backend.so /nakama/data/modules/
COPY local.yml /nakama/data/
