FROM --platform=$BUILDPLATFORM golang:1.26 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /movie-log-api ./cmd/movie-log-api

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /movie-log-api /movie-log-api
ENTRYPOINT ["/movie-log-api"]
