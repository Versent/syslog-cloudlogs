package main

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/versent/syslog-cloudlogs/pkg/batching"
	"github.com/versent/syslog-cloudlogs/pkg/config"
	"github.com/versent/syslog-cloudlogs/pkg/cwlogs"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

const (
	batchSize     = 900000
	batchDuration = 250 * time.Millisecond
)

func main() {
	var c config.SyslogConfig

	err := envconfig.Process("syslog", &c)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	err = c.Validate()
	if err != nil {
		logrus.Fatal(err.Error())
	}

	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	dispatcher, err := cwlogs.NewDispatcher(&c)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	err = dispatcher.SetupCloudwatch()
	if err != nil {
		logrus.Fatal(err.Error())
	}

	batcher := batching.NewBatcher(batchSize, batchDuration, dispatcher.Dispatch)

	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(handler)
	server.SetTlsPeerNameFunc(tlsPeerFunc)

	err = setupTLSListener(&c, server)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	err = server.Boot()
	if err != nil {
		logrus.Fatal(err.Error())
	}

	go batcher.Handler(channel)

	server.Wait()
}

func setupTLSListener(conf *config.SyslogConfig, server *syslog.Server) error {

	cert, err := conf.Certificate()
	if err != nil {
		return errors.Wrap(err, "failed to build certs from configuration")
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	addr := fmt.Sprintf(":%v", conf.Port)

	err = server.ListenTCPTLS(addr, config)
	if err != nil {
		return errors.Wrap(err, "failed to start TLS listener")
	}

	logrus.WithField("addr", addr).Info("TLS listen")

	return nil
}

func tlsPeerFunc(tlsConn *tls.Conn) (tlsPeer string, ok bool) {
	return "default", true
}
