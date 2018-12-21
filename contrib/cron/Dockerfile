FROM golang:1.11-alpine AS minicron
RUN apk add --no-cache git
ENV CGO_ENABLED 0
RUN go get -u github.com/networkteam/minicron && \
    cp "${GOPATH}/bin/minicron" /app

FROM golang:1.11-alpine AS build
COPY . "${GOPATH}"/src/app/
ENV CGO_ENABLED 0
RUN go install app && \
    cp "${GOPATH}/bin/app" /app

FROM alpine:3.7
RUN apk add --no-cache ca-certificates

COPY --from=minicron /app /minicron
COPY --from=build /app /slack-emoji-upload

VOLUME [ "/emoji" ]
RUN >/task printf "%s\n" "#!/bin/sh -eu" && \
    >>/task printf "%s\n" '/slack-emoji-upload $* /emoji/*.*' && \
    chmod +x /task
RUN >/entrypoint printf "%s\n" "#!/bin/sh -eu" && \
    >>/entrypoint printf "%s\n" 'SCHEDULE=$1' && \
    >>/entrypoint printf "%s\n" 'shift' && \
    >>/entrypoint printf "%s\n" '/minicron -v -c emoji "$SCHEDULE" "/task $*"' && \
    chmod +x /entrypoint

# Example usage:
#   docker run -v ~/Emoji:/emoji slack-emoji-upload-cron "@every 12h" -team my-team -token "$SLACK_API_TOKEN"
ENTRYPOINT [ "/entrypoint" ]
