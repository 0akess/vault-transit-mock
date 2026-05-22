FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/vault-transit-mock .

FROM scratch
COPY --from=build /out/vault-transit-mock /vault-transit-mock
EXPOSE 8200
HEALTHCHECK --interval=5s --timeout=3s --start-period=2s --retries=10 \
    CMD ["/vault-transit-mock", "healthcheck"]
ENTRYPOINT ["/vault-transit-mock"]
