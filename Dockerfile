FROM golang:1.27.1-alpine AS build

WORKDIR /app
COPY . .
RUN go build -o bot main.go

FROM alpine:3.24
WORKDIR /app
COPY --from=build /app/bot .
CMD ["./bot"]

