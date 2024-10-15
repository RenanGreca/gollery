FROM --platform=$BUILDPLATFORM golang:alpine as build

ARG VERSION
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

RUN apk add build-base

COPY . .

RUN go mod download

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate

RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o gollery ./

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/gollery .
COPY ./assets ./assets

EXPOSE 3000

ENTRYPOINT ["/app/gollery"]
