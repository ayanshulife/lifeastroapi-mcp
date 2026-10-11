# Container build of lifeastro-mcp (stdio MCP server). Used by MCP
# registries such as Smithery that run servers in a container; local
# users should prefer Homebrew, the release binaries or `go install`.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
ARG VERSION=0.11.0
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/lifeastro-mcp ./cmd/lifeastro-mcp

FROM alpine:3.20
# Ownership label checked by the official MCP Registry (must equal "name" in server.json).
LABEL io.modelcontextprotocol.server.name="io.github.ayanshulife/lifeastroapi-mcp"
LABEL org.opencontainers.image.source="https://github.com/ayanshulife/lifeastroapi-mcp"
LABEL org.opencontainers.image.description="Vedic astrology, panchang and Hindu festival calendar MCP server (LifeAstroAPI)"
LABEL org.opencontainers.image.licenses="MIT"
RUN apk add --no-cache ca-certificates && adduser -D -H mcp
COPY --from=build /out/lifeastro-mcp /usr/local/bin/lifeastro-mcp
USER mcp
ENTRYPOINT ["lifeastro-mcp"]
