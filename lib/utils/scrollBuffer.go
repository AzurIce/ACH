package utils

import "strings"

// ScrollBuffer
type ScrollBuffer struct {
	buffer [1024]string
	cursor int
}

func NewScrollBuffer() *ScrollBuffer {
	return &ScrollBuffer{}
}

func (buf *ScrollBuffer) Write(str string) {
	buf.buffer[buf.cursor] = str
	buf.cursor = (buf.cursor + 1) % 1024
}

func (buf *ScrollBuffer) Writeln(str string) {
	buf.Write(str + "\n")
}

func (buf *ScrollBuffer) GetBuf() string {
	return strings.Join(buf.buffer[buf.cursor:], "") + strings.Join(buf.buffer[:buf.cursor], "")
}

// ChanBuffer
// type ChanBuffer struct {
// 	buffer  ScrollBuffer
// 	channel chan string
// }

// func NewChanBuffer(channel chan string) *ChanBuffer {
// 	return &ChanBuffer{buffer: *NewScrollBuffer(), channel: channel}
// }

// func (buf *ChanBuffer) Writeln(str string) {
// 	buf.buffer.Writeln(str)
// 	buf.channel <- str
// }

// func (buf *ChanBuffer) GetBuf() string {
// 	return buf.buffer.GetBuf()
// }