# ---- build stage ----
FROM golang:1.25-alpine AS build
WORKDIR /src

# Download dependencies first (cached unless go.mod/go.sum change).
COPY go.mod go.sum ./
RUN go mod download

# Build both binaries: the service and the event subscriber.
COPY . .
RUN CGO_ENABLED=0 go build -o /out/todo-service .
RUN CGO_ENABLED=0 go build -o /out/subscriber ./subscriber

# ---- run stage ----
FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/todo-service /app/todo-service
COPY --from=build /out/subscriber /app/subscriber

ENV NATS_URL=nats://nats:4222

# Default command runs the service. The compose "subscriber" service
# overrides this with /app/subscriber.
CMD ["/app/todo-service"]
