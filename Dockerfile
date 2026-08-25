FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/sky-snapshot ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/sky-snapshot /sky-snapshot
USER 65532:65532
EXPOSE 8080
ENV BIND_ADDR=:8080
ENTRYPOINT ["/sky-snapshot"]
