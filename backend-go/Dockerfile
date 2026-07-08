# Build stage — compiles a static single binary.
FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Runtime stage — tiny final image (~15MB vs ~200MB+ for the JVM).
FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /app/server .
# migrations/ copied when it exists (added in the auth phase); harmless if absent.
COPY --from=build /app/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
