FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /sunpathd ./cmd/sunpathd && \
    CGO_ENABLED=0 go build -o /worker ./cmd/worker && \
    CGO_ENABLED=0 go build -o /migratedb ./cmd/migratedb

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /sunpathd /usr/local/bin/sunpathd
COPY --from=build /worker /usr/local/bin/worker
COPY --from=build /migratedb /usr/local/bin/migratedb
EXPOSE 8080
CMD ["sunpathd"]
