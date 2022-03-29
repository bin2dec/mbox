package mbox

import (
	"bufio"
	"bytes"
	"testing"
)

func TestScanMessages(t *testing.T) {
	messageList := [][]byte{
		[]byte(`From MAILER-DAEMON Fri Jul  8 12:08:34 2011
From: Author <author@example.com>
To: Recipient <recipient@example.com>
Subject: Sample message 1
 
This is the body.
>From (should be escaped).
There are 3 lines.`),
		[]byte(`From MAILER-DAEMON Fri Jul  8 12:08:34 2011
From: Author <author@example.com>
To: Recipient <recipient@example.com>
Subject: Sample message 2
 
This is the second body.`),
	}
	messages := bytes.Join(messageList, []byte("\n"))
	scanner := bufio.NewScanner(bytes.NewReader(messages))
	scanner.Split(ScanMessages)

	for i, message := range messageList {
		if !scanner.Scan() || bytes.Compare(message, scanner.Bytes()) != 0 {
			t.Errorf("#%v: got:\n%#v\nwant:\n%#v", i, scanner.Bytes(), message)
		}
	}
}
