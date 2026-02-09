FROM golang:1.25.5-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /backend ./cmd
RUN go build -o /worker ./worker/cmd

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=build /backend .
COPY --from=build /worker ./worker

EXPOSE 8080

CMD ["./backend"]
