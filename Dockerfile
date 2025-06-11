# build stage
FROM golang:1.24.2-alpine AS builder

WORKDIR /app

#sqlite deps
RUN apk add --no-cache git gcc g++ make sqlite sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN ls .

RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd

#sqlite initializer
RUN CGO_ENABLED=1 GOOS=linux go build -o db-init ./scripts/default_sqlite.go

# runtime stage
FROM alpine

WORKDIR /

COPY --from=builder /app/server .
COPY --from=builder /app/db-init .
COPY /specs/api ./specs/api

# entrypoint script
COPY --from=builder /app/scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# mount volume
VOLUME ["/data"]

# list all specs
RUN date && ls . && ls ./specs -R

EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]