FROM golang:latest AS build

WORKDIR /build
COPY go.* .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o consumer ./...


FROM scratch
WORKDIR /app
COPY --from=build /build/consumer /app/consumer
CMD ["./consumer"]
