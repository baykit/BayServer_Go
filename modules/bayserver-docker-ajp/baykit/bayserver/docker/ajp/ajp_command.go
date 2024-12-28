package ajp

type AjpCommand interface {
	ToServer() bool
	SetToServer(toServer bool)
}
