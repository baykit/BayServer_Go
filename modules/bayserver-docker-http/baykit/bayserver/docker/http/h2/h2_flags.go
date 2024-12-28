package h2

import "strconv"

const FLAGS_NONE = 0x0
const FLAGS_ACK = 0x1
const FLAGS_END_STREAM = 0x1
const FLAGS_END_HEADERS = 0x4
const FLAGS_PADDED = 0x8
const FLAGS_PRIORITY = 0x20

type H2Flags struct {
	Flags int
}

func NewH2Flags(flags int) *H2Flags {
	return &H2Flags{Flags: flags}
}

func (f *H2Flags) String() string {
	return strconv.Itoa(f.Flags)
}

func (f *H2Flags) HasFlag(flag int) bool {
	return (f.Flags & flag) != 0
}

func (f *H2Flags) SetFlag(flag int, val bool) {
	if val {
		f.Flags |= flag
	} else {
		f.Flags &= ^flag
	}
}

func (f *H2Flags) IsAck() bool {
	return f.HasFlag(FLAGS_ACK)
}

func (f *H2Flags) SetAck(isAck bool) {
	f.SetFlag(FLAGS_ACK, isAck)
}

func (f *H2Flags) IsEndStream() bool {
	return f.HasFlag(FLAGS_END_STREAM)
}

func (f *H2Flags) SetEndStream(isEndStream bool) {
	f.SetFlag(FLAGS_END_STREAM, isEndStream)
}

func (f *H2Flags) IsEndHeaders() bool {
	return f.HasFlag(FLAGS_END_HEADERS)
}

func (f *H2Flags) SetEndHeaders(isEndHeaders bool) {
	f.SetFlag(FLAGS_END_HEADERS, isEndHeaders)
}

func (f *H2Flags) IsPadded() bool {
	return f.HasFlag(FLAGS_PADDED)
}

func (f *H2Flags) SetPadded(isPadded bool) {
	f.SetFlag(FLAGS_PADDED, isPadded)
}

func (f *H2Flags) IsPriority() bool {
	return f.HasFlag(FLAGS_PRIORITY)
}

func (f *H2Flags) SetPriority(isPriority bool) {
	f.SetFlag(FLAGS_PRIORITY, isPriority)
}
