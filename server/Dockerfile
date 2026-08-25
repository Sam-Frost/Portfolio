FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/domain ./cmd

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=build /out/domain /usr/local/bin/domain
EXPOSE 8080
ENTRYPOINT ["domain"]
