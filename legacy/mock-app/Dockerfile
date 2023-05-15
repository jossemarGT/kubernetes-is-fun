FROM golang:1.20 AS builder

WORKDIR /usr/src/app

COPY go.mod ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/src/app/ ./...

FROM debian:bullseye

COPY --from=builder /usr/src/app/legacy-mock /

RUN addgroup legacy-mock \
    && adduser --system --no-create-home --ingroup legacy-mock --uid 1005 legacy-mock

USER 1005
ENTRYPOINT ["/legacy-mock"]
