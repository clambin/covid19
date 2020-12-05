FROM golang:1.15 AS builder

WORKDIR /build

COPY . ./
RUN CGO_ENABLED=0 go build cmd/covid19api/covid19api.go

FROM alpine

WORKDIR /app

COPY --from=builder /build/covid19api /app

EXPOSE 5000
ENTRYPOINT ["/app/covid19api"]
CMD []


