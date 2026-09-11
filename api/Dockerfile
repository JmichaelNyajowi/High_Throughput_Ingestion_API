FROM golang:1.26.2-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/api ./cmd/api \
  && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/seed ./cmd/seed

FROM alpine:3.21
COPY --from=build /out/api /api
COPY --from=build /out/seed /seed
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/api"]
