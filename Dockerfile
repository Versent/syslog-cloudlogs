FROM golang:1.10 as builder
ADD . /go/src/github.com/versent/syslog-cloudlogs
WORKDIR /go/src/github.com/versent/syslog-cloudlogs
RUN make setup
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o syslog-cloudlogs ./cmd/syslog-cloudlogs

FROM alpine:latest
LABEL maintainer "Mark Wolfe <mark.wolfe@versent.com.au>"
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN addgroup appuser && adduser -s /bin/bash -D -G appuser appuser
WORKDIR /app
# ADD syslog-cloudlogs /app
COPY --from=builder /go/src/github.com/versent/syslog-cloudlogs/syslog-cloudlogs .
USER appuser
ENTRYPOINT ["./syslog-cloudlogs"]