FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /controlplane ./cmd/controlplane

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /controlplane /app/controlplane
EXPOSE 9000
ENTRYPOINT ["/app/controlplane"]
