FROM golang:1.22-alpine AS builder

WORKDIR /locus

COPY go.mod go.sum ./

RUN go mod download

RUN apk update && apk add --update nodejs npm

COPY locus.go ./
COPY public public
COPY auth auth
COPY coldmfa coldmfa

WORKDIR /locus/coldmfa/app
RUN npm ci && npm run build

WORKDIR /locus

RUN go build

FROM alpine:latest

WORKDIR /app

COPY --from=builder /locus/locus /app/locus

EXPOSE 3000

CMD ["/app/locus"]
