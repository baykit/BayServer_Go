package h2

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/exception"
)

/**
 * HTTP/2 Setting payload format
 *
 * +-------------------------------+
 * |       Identifier (16)         |
 * +-------------------------------+-------------------------------+
 * |                        Value (32)                             |
 * +---------------------------------------------------------------+
 *
 */

const HEADER_TABLE_SIZE = 0x1

const ENABLE_PUSH = 0x2

const MAX_CONCURRENT_STREAMS = 0x3

const INITIAL_WINDOW_SIZE = 0x4

const MAX_FRAME_SIZE = 0x5

const MAX_HEADER_LIST_SIZE = 0x6

const INIT_HEADER_TABLE_SIZE = 4096

const INIT_ENABLE_PUSH = 1

const INIT_MAX_CONCURRENT_STREAMS = -1

const INIT_INITIAL_WINDOW_SIZE = 65535

const INIT_MAX_FRAME_SIZE = 16384

const INIT_MAX_HEADER_LIST_SIZE = -1

type cmdSettingItem struct {
	id    int
	value int
}

func NewCmdSettingItem(id int, value int) *cmdSettingItem {
	return &cmdSettingItem{
		id:    id,
		value: value,
	}
}

type CmdSettings struct {
	*H2CommandBase
	items []*cmdSettingItem
}

func NewCmdSettings(streamId int, flags *H2Flags) *CmdSettings {
	c := &CmdSettings{
		H2CommandBase: NewH2CommandBase(H2_TYPE_SETTINGS, flags, streamId),
	}

	var _ protocol.Command = c // implement check
	var _ H2Command = c        // implement check
	return c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdSettings) Unpack(pkt protocol.Packet) exception.IOException {
	c.H2CommandBase.Unpack(pkt.(*H2Packet))
	if c.Flags().IsAck() {
		return nil
	}

	acc := pkt.NewDataAccessor()
	pos := 0
	for pos < pkt.DataLen() {
		id := acc.GetShort()
		value := acc.GetInt()
		c.items = append(c.items, NewCmdSettingItem(id, value))
		pos += 6
	}
	return nil
}

func (c *CmdSettings) Pack(pkt protocol.Packet) exception.IOException {

	if c.Flags().IsAck() {
		// not pack payload

	} else {
		acc := pkt.NewDataAccessor()
		for _, item := range c.items {
			acc.PutShort(item.id)
			acc.PutInt(item.value)
		}
	}
	c.H2CommandBase.Pack(pkt.(*H2Packet))
	return nil
}

func (c *CmdSettings) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception.IOException) {
	return h.(H2CommandHandler).HandleSettings(c)
}
