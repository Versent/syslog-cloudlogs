package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	syslog "github.com/RackSec/srslog"
	"github.com/alecthomas/kingpin"
	"github.com/sirupsen/logrus"
)

var (
	// Version program version which is updated via build flags
	version = "master"

	hostaddr = kingpin.Arg("host", "Host address to connect.").Default("localhost:10514").String()
)

func main() {
	kingpin.Version(version)

	kingpin.Parse()

	logrus.Printf("connecting to %s\n", *hostaddr)

	cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client-key.pem")
	if err != nil {
		logrus.Fatal(err)
	}

	pool := x509.NewCertPool()
	serverCert, err := ioutil.ReadFile("certs/ca.pem")
	if err != nil {
		logrus.Fatal(err)
	}
	pool.AppendCertsFromPEM(serverCert)
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}
	config.BuildNameToCertificate()

	w, err := syslog.DialWithTLSConfig("tcp+tls", *hostaddr, syslog.LOG_ERR, "kernel", &config)
	if err != nil {
		logrus.Fatal(err)
	}

	err = w.Alert(`{"msg":"hello", "counter":123, "flag":true}`)
	if err != nil {
		logrus.Fatal(err)
	}

}
