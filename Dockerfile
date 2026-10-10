# Container build of lifeastro-mcp (stdio MCP server). Used by MCP
# registries such as Smithery that run servers in a container; local
# users should prefer Homebrew, the release binaries or `go install`.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
ARG VERSION=0.4.0
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/lifeastro-mcp ./cmd/lifeastro-mcp

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -H mcp
COPY --from=build /out/lifeastro-mcp /usr/local/bin/lifeastro-mcp
USER mcp
ENTRYPOINT ["lifeastro-mcp"]
