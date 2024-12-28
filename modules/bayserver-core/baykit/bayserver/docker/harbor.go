package docker

import (
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/exception"
	"strings"
)

const MULTI_PLEXER_TYPE_SPIDER int = 1
const MULTI_PLEXER_TYPE_SPIN int = 2
const MULTI_PLEXER_TYPE_PIGEON int = 3
const MULTI_PLEXER_TYPE_JOB int = 4
const MULTI_PLEXER_TYPE_TAXI int = 5
const MULTI_PLEXER_TYPE_TRAIN int = 6

const RECIPIENT_TYPE_SPIDER int = 1
const RECIPIENT_TYPE_PIPE int = 2

type Harbor interface {
	Docker

	/** Default charset */
	Charset() string

	/** Default locale */
	Locale() *util.Locale

	/** Number of grand agents */
	GrandAgents() int

	/** Number of train runners */
	TrainRunners() int

	/** Number of taxi runners */
	TaxiRunners() int

	/** Max count of ships */
	MaxShips() int

	/** Trouble base */
	Trouble() Trouble

	/** Socket timeout in seconds */
	SocketTimeoutSec() int

	/** Keep-Alive timeout in seconds */
	KeepTimeoutSec() int

	/** Trace req/res header flag */
	TraceHeader() bool

	/** Internal buffer size of Tour */
	TourBufferSize() int

	/** File name to redirect stdout/stderr */
	RedirectFile() string

	/** Port number of signal agent */
	ControlPort() int

	/** Gzip compression flag */
	GzipComp() bool

	/** NetMultiplexer of Network I/O */
	NetMultiplexer() int

	/** NetMultiplexer of File I/O */
	FileMultiplexer() int

	/** NetMultiplexer of Log output */
	LogMultiplexer() int

	/** NetMultiplexer of CGI input */
	CgiMultiplexer() int

	/** Recipient type */
	Recipient() int

	/** PID file name */
	PidFile() string

	/** Multi datalistener flag */
	MultiCore() bool
}

func GetMultiplexerTypeName(typ int) string {
	switch typ {
	case MULTI_PLEXER_TYPE_SPIDER:
		return "spider"
	case MULTI_PLEXER_TYPE_SPIN:
		return "spin"
	case MULTI_PLEXER_TYPE_PIGEON:
		return "pigeon"
	case MULTI_PLEXER_TYPE_JOB:
		return "job"
	case MULTI_PLEXER_TYPE_TAXI:
		return "taxi"
	case MULTI_PLEXER_TYPE_TRAIN:
		return "train"
	default:
		return ""
	}
}

func GetMultiPlexerType(typ string) (int, exception.Exception) {
	switch strings.ToLower(typ) {
	case "spider":
		return MULTI_PLEXER_TYPE_SPIDER, nil
	case "spin":
		return MULTI_PLEXER_TYPE_SPIN, nil
	case "pigeon":
		return MULTI_PLEXER_TYPE_PIGEON, nil
	case "job":
		return MULTI_PLEXER_TYPE_JOB, nil
	case "taxi":
		return MULTI_PLEXER_TYPE_TAXI, nil
	case "train":
		return MULTI_PLEXER_TYPE_TRAIN, nil
	default:
		return 0, exception.NewException("Illegal argument")
	}
}

func GetRecipientTypeName(typ int) string {
	switch typ {
	case RECIPIENT_TYPE_SPIDER:
		return "spider"
	case RECIPIENT_TYPE_PIPE:
		return "pipe"
	default:
		return ""
	}
}

func GetRecipientType(typ string) (int, exception.Exception) {
	switch strings.ToLower(typ) {
	case "spider":
		return RECIPIENT_TYPE_SPIDER, nil
	case "pipe":
		return RECIPIENT_TYPE_PIPE, nil
	default:
		return 0, exception.NewException("Illegal argument")
	}
}
