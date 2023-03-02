# Build
FROM golang:1.20-alpine AS build

# Install dependencies
RUN apk update && apk upgrade && apk add --no-cache \
  make git

WORKDIR /app

COPY . .

RUN make build-linux

# Final container
FROM alpine:3.17

WORKDIR /app

COPY --from=build /app/bin/linux/gitlab-code-quality /app/

RUN chmod u+x /app/gitlab-code-quality

# Start
ENTRYPOINT [ "/app/gitlab-code-quality", "transform", "--detect-report", "--output-file", "gl-code-quality-report.json" ]