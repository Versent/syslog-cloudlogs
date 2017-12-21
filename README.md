# syslog-cloudlogs

This service acts as a bridge from syslog (TLS) to cloud based logging services such as AWS cloudwatch logs and is primarily designed to enable integration of apigee logging with these services.

# features

* TLS+syslog listener
* Batched upload to AWS cloudwatch logs

# configuration

This is a 12 factor application and therefore requires a set of environment variables to be set, these are as follows.

```
# port which the service listens for TLS+syslog connections
export SYSLOG_PORT=10514
# cloudwatch group and stream to upload logs
export SYSLOG_GROUP=/versent/dev/syslog
export SYSLOG_STREAM=apigee
# These certs are base64 from the example folder
export SYSLOG_CERT=XXX
export SYSLOG_KEY=XXX
# AWS region
export AWS_REGION=ap-southeast-2
```

# certificates

Run the following code to produce a string from a PEM encoded certificate, this can then be exported in `SYSLOG_CERT` environment variable.

```
cat server.crt | base64 
```

Run the following code to produce a string from a PEM encoded key, this can then  can be exported in `SYSLOG_KEY` environment variable.

```
cat server.key | base64 
```

# todo

Things which need some work:

* peer certificate verification, this really needs to be explored as a part of testing with apigee

# License

This software is copyright Versent 2017