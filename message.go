package mbox

import (
	"bufio"
	"bytes"
	"fmt"
	"unicode"
)

var messagePrefix []byte

func init() {
	messagePrefix = []byte("From ")
}

type FromLineNotFoundError struct{}

func (e FromLineNotFoundError) Error() string {
	return `mbox: "from line" not found`
}

// MessageFromLine returns the first line of the given message without the prefix "From ".
func MessageFromLine(message []byte) (value []byte, err error) {
	message = bytes.TrimLeftFunc(message, unicode.IsSpace)
	if !bytes.HasPrefix(message, messagePrefix) {
		return nil, FromLineNotFoundError{}
	}
	message = message[len(messagePrefix):]
	if high := bytes.Index(message, []byte("\n")); high >= 0 {
		message = message[:high]
	}
	return bytes.TrimSpace(message), nil
}

type HeaderNotFoundError struct {
	Name string
}

func (e HeaderNotFoundError) Error() string {
	return fmt.Sprintf("mbox: header %q not found", e.Name)
}

// MessageHeader looks up the header with the given name in the message.
// If the header exists, the function returns the header value.
func MessageHeader(message []byte, name string) (value []byte, err error) {
	low := bytes.Index(message, []byte(name+":"))
	if low < 0 {
		return nil, HeaderNotFoundError{name}
	}
	value = message[low+len(name)+1:]
	if high := bytes.IndexByte(value, '\n'); high >= 0 {
		value = value[:high]
	}
	return bytes.TrimSpace(value), nil
}

type BodyNotFoundError struct{}

func (e BodyNotFoundError) Error() string {
	return "mbox: body not found"
}

// MessageBody returns the message body.
func MessageBody(message []byte) (body []byte, err error) {
	if low := bytes.Index(message, []byte("\n\n")); low >= 0 {
		return message[low+2:], nil
	}
	if low := bytes.Index(message, []byte("\r\n\r\n")); low >= 0 {
		return message[low+4:], nil
	}
	return nil, BodyNotFoundError{}
}

// SplitFromTo splits the value of a From/To header in two parts, name and email.
func SplitFromTo(value []byte) (name, email []byte) {
	lab := bytes.Index(value, []byte("<"))
	if lab < 0 {
		return nil, value
	}
	rab := bytes.LastIndex(value, []byte(">"))
	if rab < 0 {
		return nil, value
	}
	name = bytes.TrimSpace(value[:lab])
	email = value[lab+1 : rab]
	return name, email
}

// SplitMessage splits the given message into these parts: From_ line, headers, and body.
func SplitMessage(message []byte) (fromLine []byte, headers map[string][]byte, body []byte, err error) {
	message = bytes.TrimLeftFunc(message, unicode.IsSpace)

	if !bytes.HasPrefix(message, messagePrefix) {
		return nil, nil, nil, FromLineNotFoundError{}
	}
	message = message[len(messagePrefix):]
	sep := bytes.Index(message, []byte("\n"))
	if sep < 0 {
		return bytes.TrimSpace(message), nil, nil, nil
	}
	fromLine = bytes.TrimSpace(message[:sep])
	message = message[sep+1:]

	if sep = bytes.Index(message, []byte("\n\n")); sep >= 0 {
		body = message[sep+2:]
		message = message[:sep]
	} else if sep = bytes.Index(message, []byte("\r\n\r\n")); sep >= 0 {
		body = message[sep+4:]
		message = message[:sep]
	}

	headers = make(map[string][]byte)

	scanner := bufio.NewScanner(bytes.NewReader(message))
	for scanner.Scan() {
		header := bytes.SplitN(scanner.Bytes(), []byte(":"), 2)
		if len(header) == 2 {
			hName := string(bytes.TrimSpace(header[0]))
			hValue := bytes.TrimSpace(header[1])
			headers[hName] = hValue
		}
	}

	if len(headers) > 0 {
		return fromLine, headers, body, nil
	}
	return fromLine, nil, body, nil
}

// TrimPatchPrefix returns the value without the prefix "[PATCH...] ".
func TrimPatchPrefix(value []byte) (trimmedValue []byte) {
	trimmedValue = bytes.TrimLeftFunc(value, unicode.IsSpace)
	if bytes.HasPrefix(trimmedValue, []byte("[PATCH")) {
		if low := bytes.Index(trimmedValue, []byte("]")); low >= 0 {
			return bytes.TrimSpace(trimmedValue[low+1:])
		}
	}
	return value
}
