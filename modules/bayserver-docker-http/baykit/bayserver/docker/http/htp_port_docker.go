package http

const DEFAULT_SUPPORT_H2 = true

type HtpPortDocker interface {
	SupportH2() bool
}
