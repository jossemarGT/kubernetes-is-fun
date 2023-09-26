FROM golang:1.20 AS builder

WORKDIR /usr/src/app

COPY go.mod ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 \
    go build -v -o /usr/src/app/ --ldflags '-extldflags "-static"' ./...

FROM alpine:3.18

COPY --from=builder /usr/src/app/mock-app /

RUN addgroup mock-app \
    && adduser --system --no-create-home --ingroup mock-app --uid 1005 mock-app

USER 1005
ENTRYPOINT ["/mock-app"]
