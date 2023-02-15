
FROM golang:1.20.1-alpine AS build_deps

RUN apk add --no-cache git

WORKDIR /workspace

COPY go.mod .
COPY go.sum .

RUN go mod download

FROM build_deps AS build

COPY . .

RUN CGO_ENABLED=0 go build -o webhook -ldflags '-w -extldflags "-static"' .

FROM alpine:3.14.2

RUN apk add --no-cache ca-certificates

COPY --from=build /workspace/webhook /usr/local/bin/webhook

RUN apk add libcap && setcap 'cap_net_bind_service=+ep' /usr/local/bin/webhook

USER 1001

ENTRYPOINT ["webhook"]
