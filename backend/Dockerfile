FROM golang:1.22-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/backend ./main.go

FROM golang:1.22-alpine AS dev

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["go", "run", "./main.go"]

FROM alpine:3.20

RUN addgroup -S app && adduser -S -G app app && apk add --no-cache ca-certificates \
    && mkdir -p /frontend/public/preview/uploads \
    && chown -R app:app /frontend

WORKDIR /app

COPY --from=build /out/backend /app/backend

USER app

EXPOSE 8080

ENTRYPOINT ["/app/backend"]
