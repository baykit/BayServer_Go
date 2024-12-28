package h2

import (
	"bayserver-core/baykit/bayserver/util/exception"
	"strconv"
)

/**
 * HPack
 *   https://datatracker.ietf.org/doc/html/rfc7541
 *
 *
 */

const HEADER_OP_INDEX = 1
const HEADER_OP_OVERLOAD_KNOWN_HEADER = 2
const HEADER_OP_NEW_HEADER = 3
const HEADER_OP_KNOWN_HEADER = 4
const HEADER_OP_UNKNOWN_HEADER = 5
const HEADER_OP_UPDATE_DYNAMIC_TABLE_SIZE = 6

type HeaderBlock struct {
	op    int
	index int
	name  string
	value string
	size  int
}

func (b *HeaderBlock) String() string {
	return strconv.Itoa(b.op) + " index=" + strconv.Itoa(b.index) + " name=" + b.name + " value=" + b.value + " size=" + strconv.Itoa(b.size)
}

func NewHeaderBlock() *HeaderBlock {
	return &HeaderBlock{}
}

func HeaderBlockPack(blk *HeaderBlock, acc *H2DataAccessor) exception.IOException {
	switch blk.op {
	case HEADER_OP_INDEX:
		acc.PutHPackInt(blk.index, 7, 1)

	case HEADER_OP_OVERLOAD_KNOWN_HEADER, HEADER_OP_NEW_HEADER, HEADER_OP_UPDATE_DYNAMIC_TABLE_SIZE:
		return exception.NewIOException("Illegal state")

	case HEADER_OP_KNOWN_HEADER:
		acc.PutHPackInt(blk.index, 4, 0)
		acc.PutHPackString(blk.value, false)

	case HEADER_OP_UNKNOWN_HEADER:
		acc.PutByte(0)
		acc.PutHPackString(blk.name, false)
		acc.PutHPackString(blk.value, false)
	}

	return nil
}

func HeaderBlockUnPack(acc *H2DataAccessor) *HeaderBlock {
	blk := NewHeaderBlock()
	index := acc.GetByte()
	//BayServer.debug("index: " + index);
	indexHeaderField := (index & 0x80) != 0
	if indexHeaderField {
		// index header field
		/**
		 *   0   1   2   3   4   5   6   7
		 * +---+---+---+---+---+---+---+---+
		 * | 1 |        Index (7+)         |
		 * +---+---------------------------+
		 */
		blk.op = HEADER_OP_INDEX
		blk.index = index & 0x7F

	} else {
		// literal header field
		updateIndex := (index & 0x40) != 0
		if updateIndex {
			index = index & 0x3F
			overloadIndex := index != 0
			if overloadIndex {
				// known header name
				if index == 0x3F {
					index = index + acc.GetHPackIntRest()
				}
				blk.op = HEADER_OP_OVERLOAD_KNOWN_HEADER
				blk.index = index

				/**
				 *      0   1   2   3   4   5   6   7
				 *    +---+---+---+---+---+---+---+---+
				 *    | 0 | 1 |      Index (6+)       |
				 *    +---+---+-----------------------+
				 *    | H |     Value Length (7+)     |
				 *    +---+---------------------------+
				 *    | Value String (Length octets)  |
				 *    +-------------------------------+
				 */
				blk.value = acc.GetHPackString()

			} else {
				// new header name
				/**
				 *   0   1   2   3   4   5   6   7
				 * +---+---+---+---+---+---+---+---+
				 * | 0 | 1 |           0           |
				 * +---+---+-----------------------+
				 * | H |     Name Length (7+)      |
				 * +---+---------------------------+
				 * |  Name String (Length octets)  |
				 * +---+---------------------------+
				 * | H |     Value Length (7+)     |
				 * +---+---------------------------+
				 * | Value String (Length octets)  |
				 * +-------------------------------+
				 */
				blk.op = HEADER_OP_NEW_HEADER
				blk.name = acc.GetHPackString()
				blk.value = acc.GetHPackString()
			}

		} else {
			updateDynamicTableSize := (index & 0x20) != 0
			if updateDynamicTableSize {
				/**
				 *   0   1   2   3   4   5   6   7
				 * +---+---+---+---+---+---+---+---+
				 * | 0 | 0 | 1 |   Max size (5+)   |
				 * +---+---------------------------+
				 */
				size := index & 0x1f
				if size == 0x1f {
					size = size + acc.GetHPackIntRest()
				}
				blk.op = HEADER_OP_UPDATE_DYNAMIC_TABLE_SIZE
				blk.size = size

			} else {
				// not update index
				index = index & 0xF
				if index != 0 {
					/**
					 *   0   1   2   3   4   5   6   7
					 * +---+---+---+---+---+---+---+---+
					 * | 0 | 0 | 0 | 0 |  Index (4+)   |
					 * +---+---+-----------------------+
					 * | H |     Value Length (7+)     |
					 * +---+---------------------------+
					 * | Value String (Length octets)  |
					 * +-------------------------------+
					 *
					 * OR
					 *
					 *   0   1   2   3   4   5   6   7
					 * +---+---+---+---+---+---+---+---+
					 * | 0 | 0 | 0 | 1 |  Index (4+)   |
					 * +---+---+-----------------------+
					 * | H |     Value Length (7+)     |
					 * +---+---------------------------+
					 * | Value String (Length octets)  |
					 * +-------------------------------+
					 */
					if index == 0xF {
						index = index + acc.GetHPackIntRest()
					}
					blk.op = HEADER_OP_KNOWN_HEADER
					blk.index = index
					blk.value = acc.GetHPackString()
				} else {
					// literal header field
					/**
					 *   0   1   2   3   4   5   6   7
					 * +---+---+---+---+---+---+---+---+
					 * | 0 | 0 | 0 | 0 |       0       |
					 * +---+---+-----------------------+
					 * | H |     Name Length (7+)      |
					 * +---+---------------------------+
					 * |  Name String (Length octets)  |
					 * +---+---------------------------+
					 * | H |     Value Length (7+)     |
					 * +---+---------------------------+
					 * | Value String (Length octets)  |
					 * +-------------------------------+
					 *
					 * OR
					 *
					 *   0   1   2   3   4   5   6   7
					 * +---+---+---+---+---+---+---+---+
					 * | 0 | 0 | 0 | 1 |       0       |
					 * +---+---+-----------------------+
					 * | H |     Name Length (7+)      |
					 * +---+---------------------------+
					 * |  Name String (Length octets)  |
					 * +---+---------------------------+
					 * | H |     Value Length (7+)     |
					 * +---+---------------------------+
					 * | Value String (Length octets)  |
					 * +-------------------------------+
					 */
					blk.op = HEADER_OP_UNKNOWN_HEADER
					blk.name = acc.GetHPackString()
					blk.value = acc.GetHPackString()
				}
			}
		}
	}
	return blk

}
