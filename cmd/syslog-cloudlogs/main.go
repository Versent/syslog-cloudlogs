package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/versent/syslog-cloudlogs/pkg/batching"
	"github.com/versent/syslog-cloudlogs/pkg/config"
	"github.com/versent/syslog-cloudlogs/pkg/cwlogs"
	syslog "github.com/wolfeidau/go-syslog"
	"github.com/wolfeidau/proxyv2"
)

const (
	batchSize     = 900000
	batchDuration = 250 * time.Millisecond
)

func main() {
	var c config.SyslogConfig

	logrus.SetFormatter(&logrus.JSONFormatter{})

	err := envconfig.Process("syslog", &c)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	err = c.Validate()
	if err != nil {
		logrus.Fatal(err.Error())
	}

	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
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

	ln, err := buildListener(conf)
	if err != nil {
		return errors.Wrap(err, "failed to create listener")
	}

	cert, err := conf.Certificate()
	if err != nil {
		return errors.Wrap(err, "failed to build certs from configuration")
	}

	caCert, err := conf.ClientCaCertificate()
	if err != nil {
		return errors.Wrap(err, "failed to build ca cert from configuration")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},        // server certificate which is validated by the client
		ClientCAs:    caCertPool,                     // used to verify the client cert is signed by the CA and is therefore valid
		ClientAuth:   tls.RequireAndVerifyClientCert, // this requires a valid client certificate to be supplied during handshake
	}

	tlsLn := tls.NewListener(ln, config)

	err = server.Listen(tlsLn)
	if err != nil {
		return errors.Wrap(err, "failed to start TLS listener")
	}

	return nil
}

func buildListener(conf *config.SyslogConfig) (net.Listener, error) {

	addr := fmt.Sprintf(":%v", conf.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create TCP listener")
	}

	logrus.WithField("addr", addr).Info("TLS listen")

	// if we aren't proxying just return ln
	if !conf.Proxy {
		return ln, nil
	}

	proxyLn, err := proxyv2.NewListener(ln, &proxyv2.Config{
		Trace:            traceProxyHeaders,
		ProxyHeaderError: proxyHeaderError,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create proxy listener")
	}

	return proxyLn, nil
}

func tlsPeerFunc(tlsConn *tls.Conn) (tlsPeer string, ok bool) {
	return "default", true
}

func traceProxyHeaders(state *proxyv2.ProxyConn) {

	// skip logging health checks from NLB
	if state.WriteCounter() == 0 {
		return
	}

	logrus.WithField("proxy", map[string]interface{}{
		"source":       state.Info().V4Addr.SourceIP().String(),
		"destination":  state.Info().V4Addr.DestIP().String(),
		"bytesRead":    state.ReadCounter(),
		"bytesWritten": state.WriteCounter(),
		"TLVs":         fmt.Sprintf("%s", state.Info().TLVs),
	}).Info("trace connection")
}

func proxyHeaderError(err error) {
	logrus.WithField("proxy", map[string]interface{}{
		"error": err,
	}).Info("proxy header")
}
