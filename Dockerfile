# build stage
FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN ls .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd

# runtime stage
FROM alpine

WORKDIR /

COPY --from=builder /app/server .
COPY /specs/api ./specs/api

RUN date && ls . && ls ./specs -R

EXPOSE 8080
ENTRYPOINT ["/server"]