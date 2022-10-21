FROM golang:1.18-alpine3.16 AS builder

RUN apk add git
ADD . /src/app
WORKDIR /src/app
RUN go mod download

RUN go build -o service ./cmd/service/

FROM alpine:edge
COPY --from=builder /src/app/service /service
RUN chmod +x ./service

ENTRYPOINT ["/service"]