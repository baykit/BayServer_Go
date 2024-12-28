package impl

import (
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/protocol/protocolhandlerstore"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/strutil"
	"bayserver-docker-http/baykit/bayserver/docker/http"
	"bayserver-docker-http/baykit/bayserver/docker/http/h1"
	"bayserver-docker-http/baykit/bayserver/docker/http/h2"
	"bayserver-docker-http/baykit/bayserver/docker/http/h2/h2_error_code"
	"strings"
)

const DEFAULT_SUPPORT_H2 = true

var registerd = false

type HtpPortDockerImpl struct {
	*base.PortBase
	supportH2 bool
}

func NewHtpPort() docker.Port {
	if !registerd {
		registerProtocols()
	}
	h := &HtpPortDockerImpl{}
	h.PortBase = base.NewPortBase(h)
	h.supportH2 = DEFAULT_SUPPORT_H2

	// interface check
	var _ docker.Docker = h
	var _ docker.Port = h
	var _ http.HtpPortDocker = h
	var _ base.PortSub = h
	return h
}

func (d *HtpPortDockerImpl) String() string {
	return "HtpPortDocker"
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (d *HtpPortDockerImpl) Init(elm *bcf.BcfElement, parent docker.Docker) exception.ConfigException {
	err := d.PortBase.Init(elm, parent)
	if err != nil {
		return err
	}

	if d.supportH2 {
		if d.Secure() {
			d.PortBase.SecureDocker.SetAppProtocols([]string{"h2", "http/1.1"})
		}
		perr := h2_error_code.Init()
		if perr != nil {
			return perr
		}
	}
	//H2ErrorCode.Init()
	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (d *HtpPortDockerImpl) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception.ConfigException) {

	switch strings.ToLower(kv.Key) {
	case "supporth2", "enableh2":
		var err error
		d.supportH2, err = strutil.ParseBool(kv.Value)
		if err != nil {
			return false, exception.NewConfigException(kv.FileName, kv.LineNo, baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE, kv.Value))
		}

	default:
		return d.PortBase.InitKeyVal(kv)
	}

	return true, nil
}

/****************************************/
/* Implements Port                      */
/****************************************/

func (d *HtpPortDockerImpl) Protocol() string {
	return http.H1_PROTO_NAME
}

/****************************************/
/* Implements PortSub                  */
/****************************************/

func (d *HtpPortDockerImpl) SupportAnchored() bool {
	return true
}

func (d *HtpPortDockerImpl) SupportUnanchored() bool {
	return false
}

/****************************************/
/* Implements HtpPortDocker             */
/****************************************/

func (d *HtpPortDockerImpl) SupportH2() bool {
	return d.supportH2
}

/****************************************/
/* Static function                      */
/****************************************/

func registerProtocols() {
	packetstore.RegisterPacketProtocol(
		http.H1_PROTO_NAME,
		h1.H1PacketFactory,
	)
	packetstore.RegisterPacketProtocol(
		http.H2_PROTO_NAME,
		h2.H2PacketFactory,
	)
	protocolhandlerstore.RegisterProtocol(
		http.H1_PROTO_NAME,
		true,
		h1.H1InboundProtocolHandlerFactory,
	)
	protocolhandlerstore.RegisterProtocol(
		http.H2_PROTO_NAME,
		true,
		h2.H2InboundProtocolHandlerFactory,
	)
	registerd = true
}
