FROM alpine
LABEL maintainer "Mark Wolfe <mark.wolfe@versent.com.au>"

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN addgroup appuser && adduser -s /bin/bash -D -G appuser appuser
WORKDIR /app
ADD syslog-cloudlogs /app

USER appuser

ENTRYPOINT ["./syslog-cloudlogs"]
