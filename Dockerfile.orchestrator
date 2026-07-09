FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /orchestrator .

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /orchestrator /app/orchestrator
COPY dag.yaml /app/dag.yaml
EXPOSE 8000
ENTRYPOINT ["/app/orchestrator"]
