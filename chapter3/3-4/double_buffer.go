package main

import (
	"log"
	"io"
)

type DoubleBuffer struct {
	buf         [][]byte
	bufSize     int
	lexemeBegin int
	forward     int
	curBuf      int
	isCross     bool
	inputSrc    io.Reader
}

const EOF byte = 128
func newDoubleBuffer(bufSize int, inputSrc io.Reader) *DoubleBuffer {
	if bufSize <= 0 || inputSrc == nil {
		log.Print("newDoubleBuffer(): bufSize == ", bufSize, ", 8inputSrc == ", inputSrc)
		return nil
	}
	df := &DoubleBuffer{buf: make([][]byte, 2), bufSize: bufSize}
	df.buf[0] = make([]byte, bufSize)
	df.buf[1] = make([]byte, bufSize)
	df.buf[0][bufSize - 1] = EOF // sentinel
	df.buf[1][bufSize - 1] = EOF // sentinel
	df.curBuf = 0
	df.lexemeBegin = 0
	df.forward = df.lexemeBegin
	df.isCross = false
	df.inputSrc = inputSrc

	n, err := df.inputSrc.Read(df.buf[df.curBuf][:bufSize])
	if err != nil && err != io.EOF {
		log.Fatalln("newDoubleBuffer():", err)
	}
	df.buf[df.curBuf][n] = EOF
	return df
}

func (df *DoubleBuffer) nextLexeme() string {
	if !df.isCross {
		lexeme := string(df.buf[df.curBuf][df.lexemeBegin:df.forward])
		df.lexemeBegin = df.forward
		return lexeme
	} else {
		part1 := string(df.buf[df.curBuf][df.lexemeBegin:df.bufSize])
		df.curBuf = (df.curBuf + 1) % 2
		part2 := string(df.buf[df.curBuf][:df.forward])
		return part1 + part2
	}
}

func (df *DoubleBuffer) nextChar() (byte, error) {
	ch := df.buf[df.curBuf][df.forward]
	switch ch {
	case EOF:
		if !df.isCross && df.forward - df.curBuf == df.bufSize - 1 { // forward is at the end of a buffer
			another := (df.curBuf + 1) % 2
			// load rest input into another buffer
			n, err := df.inputSrc.Read(df.buf[another][:df.bufSize])
			if err != nil && err != io.EOF {
				log.Fatalln("DoubleBuffer::scan(): load rest input into another buffer,", err)
			}
			df.buf[another][n] = EOF
			df.isCross = true
			df.forward = 0
			ch = df.buf[another][df.forward]
		} else { // forward is at the end of input
			log.Println("DoubleBuffer::scan(): forward is at the end of input.")
			return EOF, io.EOF
		}
	default:
		df.forward++
	}
	return ch, nil
}

func (df *DoubleBuffer) backword() {
	if df.isCross {
		if df.forward != 0 {
			df.forward--
		} else {
			df.isCross = false
			df.forward = df.bufSize - 1
		}
	} else {
		df.forward--
	}
}