# syslog-cloudlogs

This service acts as a bridge from syslog to cloud based logging services such as AWS cloudwatch logs and is primarily designed to enable integration of apigee logging with these services.

# features

* TLS+syslog listener
* Batched upload to AWS cloudwatch logs
* support for [AWS NLB proxy protocol v2](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-target-groups.html#proxy-protocol)

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
# Enable proxy protocol v2 support for NLB
export SYSLOG_PROXY=true
```

# Generate self-signed certificates

CloudFlare's distributes [cfssl](https://github.com/cloudflare/cfssl) source code on github page and binaries on cfssl website.

Our documentation assumes that you will run cfssl on your local x86_64 Linux or OSX host.


```
curl -s -L -o /usr/local/bin/cfssl https://pkg.cfssl.org/R1.2/cfssl_linux-amd64
curl -s -L -o /usr/local/bin/cfssljson https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
chmod +x /usr/local/bin/{cfssl,cfssljson}
```

Initialize a certificate authority
First of all we have to save default cfssl options for future substitutions:

```
mkdir ~/cfssl
cd ~/cfssl
cfssl print-defaults config > ca-config.json
cfssl print-defaults csr > ca-csr.json
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

This code is released under MIT License.