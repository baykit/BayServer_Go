package huffman

import "bytes"

var HTreeRoot = NewHNode()
var initialized = false

func HTreeDecode(data []byte) string {
	w := bytes.Buffer{}
	cur := HTreeRoot
	for _, c := range data {
		for j := 0; j < 8; j++ {
			bit := c >> (8 - j - 1) & 0x1

			// down tree
			if bit == 1 {
				cur = cur.One
			} else {
				cur = cur.Zero
			}

			if cur.Value > 0 {
				// leaf node
				w.Write([]byte{byte(cur.Value)})
				cur = HTreeRoot
			}
		}
	}
	return w.String()
}

func hTreeInsert(code int, lenInBits int, sym int) {
	bits := make([]int, lenInBits)
	for i := range bits {
		bits[i] = code >> (lenInBits - i - 1) & 0x01
	}
	hTreeInsertBits(bits, sym)
}

func hTreeInsertBits(codes []int, sym int) {
	curNode := HTreeRoot
	for _, c := range codes {
		if c == 1 {
			if curNode.One == nil {
				curNode.One = NewHNode()
			}
			curNode = curNode.One

		} else {
			if curNode.Zero == nil {
				curNode.Zero = NewHNode()
			}
			curNode = curNode.Zero
		}
	}
	curNode.Value = sym
}

func HTreeInit() bool {
	if initialized {
		return true
	}

	hTreeInsert(0x1ff8, 13, 0)
	hTreeInsert(0x7fffd8, 23, 1)
	hTreeInsert(0xfffffe2, 28, 2)
	hTreeInsert(0xfffffe3, 28, 3)
	hTreeInsert(0xfffffe4, 28, 4)
	hTreeInsert(0xfffffe5, 28, 5)
	hTreeInsert(0xfffffe6, 28, 6)
	hTreeInsert(0xfffffe7, 28, 7)
	hTreeInsert(0xfffffe8, 28, 8)
	hTreeInsert(0xffffea, 24, 9)
	hTreeInsert(0x3ffffffc, 30, 10)
	hTreeInsert(0xfffffe9, 28, 11)
	hTreeInsert(0xfffffea, 28, 12)
	hTreeInsert(0x3ffffffd, 30, 13)
	hTreeInsert(0xfffffeb, 28, 14)
	hTreeInsert(0xfffffec, 28, 15)
	hTreeInsert(0xfffffed, 28, 16)
	hTreeInsert(0xfffffee, 28, 17)
	hTreeInsert(0xfffffef, 28, 18)
	hTreeInsert(0xffffff0, 28, 19)
	hTreeInsert(0xffffff1, 28, 20)
	hTreeInsert(0xffffff2, 28, 21)
	hTreeInsert(0x3ffffffe, 30, 22)
	hTreeInsert(0xffffff3, 28, 23)
	hTreeInsert(0xffffff4, 28, 24)
	hTreeInsert(0xffffff5, 28, 25)
	hTreeInsert(0xffffff6, 28, 26)
	hTreeInsert(0xffffff7, 28, 27)
	hTreeInsert(0xffffff8, 28, 28)
	hTreeInsert(0xffffff9, 28, 29)
	hTreeInsert(0xffffffa, 28, 30)
	hTreeInsert(0xffffffb, 28, 31)
	hTreeInsert(0x14, 6, 32)
	hTreeInsert(0x3f8, 10, 33)
	hTreeInsert(0x3f9, 10, 34)
	hTreeInsert(0xffa, 12, 35)
	hTreeInsert(0x1ff9, 13, 36)
	hTreeInsert(0x15, 6, 37)
	hTreeInsert(0xf8, 8, 38)
	hTreeInsert(0x7fa, 11, 39)
	hTreeInsert(0x3fa, 10, 40)
	hTreeInsert(0x3fb, 10, 41)
	hTreeInsert(0xf9, 8, 42)
	hTreeInsert(0x7fb, 11, 43)
	hTreeInsert(0xfa, 8, 44)
	hTreeInsert(0x16, 6, 45)
	hTreeInsert(0x17, 6, 46)
	hTreeInsert(0x18, 6, 47)
	hTreeInsert(0x0, 5, 48)
	hTreeInsert(0x1, 5, 49)
	hTreeInsert(0x2, 5, 50)
	hTreeInsert(0x19, 6, 51)
	hTreeInsert(0x1a, 6, 52)
	hTreeInsert(0x1b, 6, 53)
	hTreeInsert(0x1c, 6, 54)
	hTreeInsert(0x1d, 6, 55)
	hTreeInsert(0x1e, 6, 56)
	hTreeInsert(0x1f, 6, 57)
	hTreeInsert(0x5c, 7, 58)
	hTreeInsert(0xfb, 8, 59)
	hTreeInsert(0x7ffc, 15, 60)
	hTreeInsert(0x20, 6, 61)
	hTreeInsert(0xffb, 12, 62)
	hTreeInsert(0x3fc, 10, 63)
	hTreeInsert(0x1ffa, 13, 64)
	hTreeInsert(0x21, 6, 65)
	hTreeInsert(0x5d, 7, 66)
	hTreeInsert(0x5e, 7, 67)
	hTreeInsert(0x5f, 7, 68)
	hTreeInsert(0x60, 7, 69)
	hTreeInsert(0x61, 7, 70)
	hTreeInsert(0x62, 7, 71)
	hTreeInsert(0x63, 7, 72)
	hTreeInsert(0x64, 7, 73)
	hTreeInsert(0x65, 7, 74)
	hTreeInsert(0x66, 7, 75)
	hTreeInsert(0x67, 7, 76)
	hTreeInsert(0x68, 7, 77)
	hTreeInsert(0x69, 7, 78)
	hTreeInsert(0x6a, 7, 79)
	hTreeInsert(0x6b, 7, 80)
	hTreeInsert(0x6c, 7, 81)
	hTreeInsert(0x6d, 7, 82)
	hTreeInsert(0x6e, 7, 83)
	hTreeInsert(0x6f, 7, 84)
	hTreeInsert(0x70, 7, 85)
	hTreeInsert(0x71, 7, 86)
	hTreeInsert(0x72, 7, 87)
	hTreeInsert(0xfc, 8, 88)
	hTreeInsert(0x73, 7, 89)
	hTreeInsert(0xfd, 8, 90)
	hTreeInsert(0x1ffb, 13, 91)
	hTreeInsert(0x7fff0, 19, 92)
	hTreeInsert(0x1ffc, 13, 93)
	hTreeInsert(0x3ffc, 14, 94)
	hTreeInsert(0x22, 6, 95)
	hTreeInsert(0x7ffd, 15, 96)
	hTreeInsert(0x3, 5, 97)
	hTreeInsert(0x23, 6, 98)
	hTreeInsert(0x4, 5, 99)
	hTreeInsert(0x24, 6, 100)
	hTreeInsert(0x5, 5, 101)
	hTreeInsert(0x25, 6, 102)
	hTreeInsert(0x26, 6, 103)
	hTreeInsert(0x27, 6, 104)
	hTreeInsert(0x6, 5, 105)
	hTreeInsert(0x74, 7, 106)
	hTreeInsert(0x75, 7, 107)
	hTreeInsert(0x28, 6, 108)
	hTreeInsert(0x29, 6, 109)
	hTreeInsert(0x2a, 6, 110)
	hTreeInsert(0x7, 5, 111)
	hTreeInsert(0x2b, 6, 112)
	hTreeInsert(0x76, 7, 113)
	hTreeInsert(0x2c, 6, 114)
	hTreeInsert(0x8, 5, 115)
	hTreeInsert(0x9, 5, 116)
	hTreeInsert(0x2d, 6, 117)
	hTreeInsert(0x77, 7, 118)
	hTreeInsert(0x78, 7, 119)
	hTreeInsert(0x79, 7, 120)
	hTreeInsert(0x7a, 7, 121)
	hTreeInsert(0x7b, 7, 122)
	hTreeInsert(0x7ffe, 15, 123)
	hTreeInsert(0x7fc, 11, 124)
	hTreeInsert(0x3ffd, 14, 125)
	hTreeInsert(0x1ffd, 13, 126)
	hTreeInsert(0xffffffc, 28, 127)
	hTreeInsert(0xfffe6, 20, 128)
	hTreeInsert(0x3fffd2, 22, 129)
	hTreeInsert(0xfffe7, 20, 130)
	hTreeInsert(0xfffe8, 20, 131)
	hTreeInsert(0x3fffd3, 22, 132)
	hTreeInsert(0x3fffd4, 22, 133)
	hTreeInsert(0x3fffd5, 22, 134)
	hTreeInsert(0x7fffd9, 23, 135)
	hTreeInsert(0x3fffd6, 22, 136)
	hTreeInsert(0x7fffda, 23, 137)
	hTreeInsert(0x7fffdb, 23, 138)
	hTreeInsert(0x7fffdc, 23, 139)
	hTreeInsert(0x7fffdd, 23, 140)
	hTreeInsert(0x7fffde, 23, 141)
	hTreeInsert(0xffffeb, 24, 142)
	hTreeInsert(0x7fffdf, 23, 143)
	hTreeInsert(0xffffec, 24, 144)
	hTreeInsert(0xffffed, 24, 145)
	hTreeInsert(0x3fffd7, 22, 146)
	hTreeInsert(0x7fffe0, 23, 147)
	hTreeInsert(0xffffee, 24, 148)
	hTreeInsert(0x7fffe1, 23, 149)
	hTreeInsert(0x7fffe2, 23, 150)
	hTreeInsert(0x7fffe3, 23, 151)
	hTreeInsert(0x7fffe4, 23, 152)
	hTreeInsert(0x1fffdc, 21, 153)
	hTreeInsert(0x3fffd8, 22, 154)
	hTreeInsert(0x7fffe5, 23, 155)
	hTreeInsert(0x3fffd9, 22, 156)
	hTreeInsert(0x7fffe6, 23, 157)
	hTreeInsert(0x7fffe7, 23, 158)
	hTreeInsert(0xffffef, 24, 159)
	hTreeInsert(0x3fffda, 22, 160)
	hTreeInsert(0x1fffdd, 21, 161)
	hTreeInsert(0xfffe9, 20, 162)
	hTreeInsert(0x3fffdb, 22, 163)
	hTreeInsert(0x3fffdc, 22, 164)
	hTreeInsert(0x7fffe8, 23, 165)
	hTreeInsert(0x7fffe9, 23, 166)
	hTreeInsert(0x1fffde, 21, 167)
	hTreeInsert(0x7fffea, 23, 168)
	hTreeInsert(0x3fffdd, 22, 169)
	hTreeInsert(0x3fffde, 22, 170)
	hTreeInsert(0xfffff0, 24, 171)
	hTreeInsert(0x1fffdf, 21, 172)
	hTreeInsert(0x3fffdf, 22, 173)
	hTreeInsert(0x7fffeb, 23, 174)
	hTreeInsert(0x7fffec, 23, 175)
	hTreeInsert(0x1fffe0, 21, 176)
	hTreeInsert(0x1fffe1, 21, 177)
	hTreeInsert(0x3fffe0, 22, 178)
	hTreeInsert(0x1fffe2, 21, 179)
	hTreeInsert(0x7fffed, 23, 180)
	hTreeInsert(0x3fffe1, 22, 181)
	hTreeInsert(0x7fffee, 23, 182)
	hTreeInsert(0x7fffef, 23, 183)
	hTreeInsert(0xfffea, 20, 184)
	hTreeInsert(0x3fffe2, 22, 185)
	hTreeInsert(0x3fffe3, 22, 186)
	hTreeInsert(0x3fffe4, 22, 187)
	hTreeInsert(0x7ffff0, 23, 188)
	hTreeInsert(0x3fffe5, 22, 189)
	hTreeInsert(0x3fffe6, 22, 190)
	hTreeInsert(0x7ffff1, 23, 191)
	hTreeInsert(0x3ffffe0, 26, 192)
	hTreeInsert(0x3ffffe1, 26, 193)
	hTreeInsert(0xfffeb, 20, 194)
	hTreeInsert(0x7fff1, 19, 195)
	hTreeInsert(0x3fffe7, 22, 196)
	hTreeInsert(0x7ffff2, 23, 197)
	hTreeInsert(0x3fffe8, 22, 198)
	hTreeInsert(0x1ffffec, 25, 199)
	hTreeInsert(0x3ffffe2, 26, 200)
	hTreeInsert(0x3ffffe3, 26, 201)
	hTreeInsert(0x3ffffe4, 26, 202)
	hTreeInsert(0x7ffffde, 27, 203)
	hTreeInsert(0x7ffffdf, 27, 204)
	hTreeInsert(0x3ffffe5, 26, 205)
	hTreeInsert(0xfffff1, 24, 206)
	hTreeInsert(0x1ffffed, 25, 207)
	hTreeInsert(0x7fff2, 19, 208)
	hTreeInsert(0x1fffe3, 21, 209)
	hTreeInsert(0x3ffffe6, 26, 210)
	hTreeInsert(0x7ffffe0, 27, 211)
	hTreeInsert(0x7ffffe1, 27, 212)
	hTreeInsert(0x3ffffe7, 26, 213)
	hTreeInsert(0x7ffffe2, 27, 214)
	hTreeInsert(0xfffff2, 24, 215)
	hTreeInsert(0x1fffe4, 21, 216)
	hTreeInsert(0x1fffe5, 21, 217)
	hTreeInsert(0x3ffffe8, 26, 218)
	hTreeInsert(0x3ffffe9, 26, 219)
	hTreeInsert(0xffffffd, 28, 220)
	hTreeInsert(0x7ffffe3, 27, 221)
	hTreeInsert(0x7ffffe4, 27, 222)
	hTreeInsert(0x7ffffe5, 27, 223)
	hTreeInsert(0xfffec, 20, 224)
	hTreeInsert(0xfffff3, 24, 225)
	hTreeInsert(0xfffed, 20, 226)
	hTreeInsert(0x1fffe6, 21, 227)
	hTreeInsert(0x3fffe9, 22, 228)
	hTreeInsert(0x1fffe7, 21, 229)
	hTreeInsert(0x1fffe8, 21, 230)
	hTreeInsert(0x7ffff3, 23, 231)
	hTreeInsert(0x3fffea, 22, 232)
	hTreeInsert(0x3fffeb, 22, 233)
	hTreeInsert(0x1ffffee, 25, 234)
	hTreeInsert(0x1ffffef, 25, 235)
	hTreeInsert(0xfffff4, 24, 236)
	hTreeInsert(0xfffff5, 24, 237)
	hTreeInsert(0x3ffffea, 26, 238)
	hTreeInsert(0x7ffff4, 23, 239)
	hTreeInsert(0x3ffffeb, 26, 240)
	hTreeInsert(0x7ffffe6, 27, 241)
	hTreeInsert(0x3ffffec, 26, 242)
	hTreeInsert(0x3ffffed, 26, 243)
	hTreeInsert(0x7ffffe7, 27, 244)
	hTreeInsert(0x7ffffe8, 27, 245)
	hTreeInsert(0x7ffffe9, 27, 246)
	hTreeInsert(0x7ffffea, 27, 247)
	hTreeInsert(0x7ffffeb, 27, 248)
	hTreeInsert(0xffffffe, 28, 249)
	hTreeInsert(0x7ffffec, 27, 250)
	hTreeInsert(0x7ffffed, 27, 251)
	hTreeInsert(0x7ffffee, 27, 252)
	hTreeInsert(0x7ffffef, 27, 253)
	hTreeInsert(0x7fffff0, 27, 254)
	hTreeInsert(0x3ffffee, 26, 255)
	hTreeInsert(0x3fffffff, 30, 256)

	initialized = true
	return true
}

var dummy = HTreeInit()