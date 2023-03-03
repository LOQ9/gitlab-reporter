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

COPY --from=build /app/bin/linux/gitlab-reporter /app/

RUN chmod u+x /app/gitlab-reporter

# Start
ENTRYPOINT [ "/app/gitlab-reporter" ]