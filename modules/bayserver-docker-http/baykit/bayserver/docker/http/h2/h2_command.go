package h2

type H2Command interface {
	Flags() *H2Flags
	StreamId() int
}
