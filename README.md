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
# These certs are base64 from the certs folder
export SYSLOG_CLIENTCACERT=XXX
export SYSLOG_CERT=XXX
export SYSLOG_KEY=XXX
# AWS region
export AWS_REGION=ap-southeast-2
# Enable proxy protocol v2 support for NLB
export SYSLOG_PROXY=true
# Enable debug level logging
export SYSLOG_DEBUG=true
```

# Generate self-signed certificates

CloudFlare's distributes [cfssl](https://github.com/cloudflare/cfssl) source code on github page and binaries on cfssl website.

This documentation assumes that you will run cfssl on your local x86_64 Linux or OSX host, that said you can download windows binaries from https://pkg.cfssl.org/.

```
curl -s -L -o /usr/local/bin/cfssl https://pkg.cfssl.org/R1.2/cfssl_linux-amd64
curl -s -L -o /usr/local/bin/cfssljson https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
chmod +x /usr/local/bin/{cfssl,cfssljson}
```

Navigate to provided CSR files provided.

```
cd certs
```

Generate the CA certificate and private key.

```
cfssl gencert -initca ca-csr.json | cfssljson -bare ca
```

Generate a server cert using the CSR provided.

```
cfssl gencert  \
    -ca=ca.pem \
    -ca-key=ca-key.pem \
    -config=ca-config.json \
    -hostname=localhost,127.0.0.1 \
    -profile=massl server-csr.json | cfssljson -bare server
```

Generate a client cert using the CSR provided.

```
cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=ca-config.json \
  -profile=massl \
  client-csr.json | cfssljson -bare client
```

# certificates

Run the following code to produce a string from a PEM encoded certificate, this can then be exported in `SYSLOG_CLIENTCACERT` environment variable.

```
cat certs/ca.pem | base64 
```

Run the following code to produce a string from a PEM encoded certificate, this can then be exported in `SYSLOG_CERT` environment variable.

```
cat certs/server.pem | base64 
```

Run the following code to produce a string from a PEM encoded key, this can then  can be exported in `SYSLOG_KEY` environment variable.

```
cat certs/server-key.pem | base64 
```

# License

This code is released under MIT License.