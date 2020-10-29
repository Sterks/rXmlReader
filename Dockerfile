FROM golang:latest as builder

WORKDIR /app
ADD . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o rXmlReader

FROM alpine:latest AS production
WORKDIR /app
ENV application=production
COPY --from=builder /app .
CMD ["./rXmlReader"]
