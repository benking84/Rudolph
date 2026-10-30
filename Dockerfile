FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /out/rudolph-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/rudolph-api /rudolph-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/rudolph-api"]
