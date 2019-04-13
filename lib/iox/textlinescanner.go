package iox

import (
	"bufio"
	"io"
)

type TextLineScanner struct {
	reader  *bufio.Reader
	hasNext bool
	line    string
	err     error
}

func NewTextLineScanner(reader io.Reader) *TextLineScanner {
	return &TextLineScanner{
		reader: bufio.NewReader(reader),
	}
}

func (scanner *TextLineScanner) Scan() bool {
	var err error
	scanner.line, err = scanner.reader.ReadString('\n')
	if err == io.EOF {
		return false
	} else if err != nil {
		scanner.err = err
		return false
	}
	return true
}

func (scanner *TextLineScanner) Line() (string, error) {
	return scanner.line, scanner.err
}
