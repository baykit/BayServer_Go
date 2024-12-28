package builtin

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/agent/multiplexer"
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/strutil"
	"bayserver-core/baykit/bayserver/util/sysutil"
	"crypto/tls"
	"golang.org/x/crypto/pkcs12"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type BuiltInSecureDocker struct {
	*base.DockerBase

	// SSL setting
	keyFile      string
	certFile     string
	keyStore     string
	keyStorePass string
	certs        string
	certsPass    string
	clientAuth   bool
	sslProtocol  string
	traceSSL     bool
	config       *tls.Config
}

func NewBuiltInSecureDocker() docker.Secure {
	t := &BuiltInSecureDocker{}
	t.DockerBase = base.NewDockerBase(t)

	// interface check
	var _ docker.Secure = t
	var _ docker.Docker = t
	return t
}

func (t *BuiltInSecureDocker) String() string {
	return "BuiltInSecureDocker"
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (t *BuiltInSecureDocker) Init(elm *bcf.BcfElement, parent docker.Docker) exception2.ConfigException {
	err := t.DockerBase.Init(elm, parent)
	if err != nil {
		return err
	}

	ioerr := t.initSSL()
	if ioerr != nil {
		baylog.ErrorE(ioerr, "")
		return exception2.NewConfigException(elm.FileName, elm.LineNo, "SSL init error: %s", ioerr)
	}

	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (t *BuiltInSecureDocker) InitDocker(dkr docker.Docker) (bool, exception2.ConfigException) {
	return t.DockerBase.DefaultInitDocker()
}

func (t *BuiltInSecureDocker) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception2.ConfigException) {

	var err error = nil
	var ioerr exception.IOException = nil
	for { // try catch
		switch strings.ToLower(kv.Key) {
		default:
			return t.DockerBase.DefaultInitKeyVal(kv)

		case "key":
			t.keyFile, ioerr = t.getFilePath(kv.Value)
			break

		case "cert":
			t.certFile, ioerr = t.getFilePath(kv.Value)
			break

		case "keystore":
			t.keyStore, ioerr = t.getFilePath(kv.Value)
			break

		case "keystorepass":
			t.keyStorePass, ioerr = t.getFilePath(kv.Value)
			break

		case "clientauth":
			t.clientAuth, err = strutil.ParseBool(kv.Value)
			break

		case "sslprotocol":
			t.sslProtocol = kv.Value
			break

		case "trustcerts":
			t.certs, ioerr = t.getFilePath(kv.Value)
			break

		case "certspass":
			t.certsPass = kv.Value
			break

		case "tracessl":
			t.traceSSL, err = strutil.ParseBool(kv.Value)
			break
		}

		break
	}

	if ioerr != nil {
		return false, exception2.NewConfigException(kv.FileName, kv.LineNo, ioerr.Error())
	}

	if err != nil {
		return false, exception2.NewConfigException(kv.FileName, kv.LineNo, err.Error())
	}

	return true, nil
}

/****************************************/
/* Implements Secure                    */
/****************************************/

func (t *BuiltInSecureDocker) SetAppProtocols(protocols []string) {
	t.config.NextProtos = protocols
}

func (t *BuiltInSecureDocker) ReloadCert() {
	ioerr := t.initSSL()
	if ioerr != nil {
		baylog.ErrorE(ioerr, "Reload cert error")
	}
}

func (t *BuiltInSecureDocker) NewTransporter(agtId int, sip ship.Ship) common.Transporter {
	tp := multiplexer.NewSecureTransporter(
		agent.Get(agtId).NetMultiplexer(),
		sip,
		true,
		-1,
		t.traceSSL)
	tp.Init()
	return tp
}

func (t *BuiltInSecureDocker) GetSecureConn(conn net.Conn) (net.Conn, exception.IOException) {
	tlsConn := tls.Server(conn, t.config)
	err := tlsConn.Handshake()
	if err != nil {
		return nil, exception.NewIOExceptionFromError(err)
	}

	return tlsConn, nil
}

/****************************************/
/* Private functions                    */
/****************************************/

func (t *BuiltInSecureDocker) getFilePath(path string) (string, exception.IOException) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(bayserver.BservHome(), path)
	}

	if !sysutil.IsFile(path) {
		return "", exception.NewIOException("File not found: " + path)

	} else {
		return path, nil
	}
}

func (t *BuiltInSecureDocker) initSSL() exception.IOException {

	var err error
	var cert tls.Certificate
	if t.keyStore == "" {

		cert, err = tls.LoadX509KeyPair(t.certFile, t.keyFile)
		if err != nil {
			return exception.NewIOException("Key or cert file load error: %s", err)
		}

	} else {
		// Reads PKCS#12 file
		p12Data, err := os.ReadFile(t.keyStore)
		if err != nil {
			return exception.NewIOException("failed to read pkcs12 file: %s", err)
		}

		// Decodes PKCS#12 file
		privateKey, certificate, err := pkcs12.Decode(p12Data, t.keyStorePass)
		if err != nil {
			return exception.NewIOException("Failed to decode pkcs12 file: %s", err)
		}

		// 読み取った証明書と秘密鍵をTLS証明書に変換
		cert = tls.Certificate{
			Certificate: [][]byte{certificate.Raw},
			PrivateKey:  privateKey,
		}
	}

	t.config = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		ClientAuth:         tls.NoClientCert,
		InsecureSkipVerify: true,
	}

	return nil
}
