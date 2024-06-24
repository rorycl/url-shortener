# gcr.io builder 

# setup module environment
FROM golang:1.22 AS deps
WORKDIR /build
ADD go.mod go.sum ./
RUN go mod download

# build
FROM deps as dev
COPY *go ./
COPY data ./data/
COPY templates ./templates/
COPY static ./static/
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags "-w -X main.docker=true" -o url-shortener .

# install into minimal image
FROM gcr.io/distroless/base AS base
WORKDIR /
EXPOSE 8000
COPY --from=dev /build/url-shortener /
CMD ["/webserver", "--address", "0.0.0.0", "--port", "8000"]
