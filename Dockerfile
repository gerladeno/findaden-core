FROM golang:1.18-alpine3.14 AS builder
RUN apk add git
ADD . /src/app
WORKDIR /src/app
RUN go mod download
ARG APP_BUILD_VERSION
RUN echo "Building version:  ${APP_BUILD_VERSION}"
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags " -X main.version=${APP_BUILD_VERSION}" -o homie-core ./cmd/core/

FROM alpine:edge
COPY --from=builder /src/app/homie-core /homie-core
RUN chmod +x ./homie-core
ENTRYPOINT ["/homie-core"]