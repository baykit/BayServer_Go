package h2

const DEFAULT_HEADER_TABLE_SIZE = 4096

const DEFAULT_ENABLE_PUSH = true

const DEFAULT_MAX_CONCURRENT_STREAMS = -1

const DEFAULT_MAX_WINDOW_SIZE = 65535

const DEFAULT_MAX_FRAME_SIZE = 16384

const DEFAULT_MAX_HEADER_LIST_SIZE = -1

type H2Settings struct {
	HeaderTableSize      int
	EnablePush           bool
	MaxConcurrentStreams int
	InitialWindowSize    int
	MaxFrameSize         int
	MaxHeaderListSize    int
}

func NewH2Settings() *H2Settings {
	s := &H2Settings{}
	s.Reset()
	return s
}

func (s *H2Settings) Reset() {
	s.HeaderTableSize = DEFAULT_HEADER_TABLE_SIZE
	s.EnablePush = DEFAULT_ENABLE_PUSH
	s.MaxConcurrentStreams = DEFAULT_MAX_CONCURRENT_STREAMS
	s.InitialWindowSize = DEFAULT_MAX_WINDOW_SIZE
	s.MaxFrameSize = DEFAULT_MAX_FRAME_SIZE
	s.MaxHeaderListSize = DEFAULT_MAX_HEADER_LIST_SIZE
}
