package mbox

import (
	"bytes"
	"reflect"
	"testing"
)

var testMessage = []byte(`From MAILER-DAEMON Fri Jul  8 12:08:34 2011
From: Author <author@example.com>
To: Recipient <recipient@example.com>
Subject: Sample message 1

This is the body.
>From (should be escaped).
There are 3 lines.`)

func TestMessageFromLine(t *testing.T) {
	valueWant := []byte("MAILER-DAEMON Fri Jul  8 12:08:34 2011")
	value, err := MessageFromLine(testMessage)
	if err != nil {
		t.Errorf("MessageFromLine(…) should't have returned %T (%[1]s)", err)
	} else if bytes.Compare(value, valueWant) != 0 {
		t.Errorf("MessageFromLine(…) have returned %q instead of %q", value, valueWant)
	}

	_, err = MessageFromLine(testMessage[5:])
	if _, ok := err.(FromLineNotFoundError); !ok {
		t.Errorf("MessageFromLine(…) have returned %T insted of %T", err, FromLineNotFoundError{})
	}
}

func TestMessageHeader(t *testing.T) {
	hName := "From"
	hValueWant := []byte("Author <author@example.com>")
	hValue, err := MessageHeader(testMessage, hName)
	if err != nil {
		t.Errorf("MessageHeader(…, %q) should't have returned %T (%[2]s)", hName, err)
	} else if bytes.Compare(hValue, hValueWant) != 0 {
		t.Errorf("MessageHeader(…, %q) have returned %q instead of %q", hName, hValue, hValueWant)
	}

	hName = "Date"
	_, err = MessageHeader(testMessage, hName)
	if _, ok := err.(HeaderNotFoundError); !ok {
		t.Errorf("MessageHeader(…, %q) have returned %#v insted of %#v", hName, err, HeaderNotFoundError{hName})
	}
}

func TestMessageBody(t *testing.T) {
	bodyWant := []byte("This is the body.\n>From (should be escaped).\nThere are 3 lines.")
	body, err := MessageBody(testMessage)
	if err != nil {
		t.Errorf("MessageBody(…) should't have returned %T (%[1]s)", err)
	} else if bytes.Compare(body, bodyWant) != 0 {
		t.Errorf("MessageBody(…) have returned\n%#v instead of\n%#v", body, bodyWant)
	}

	high := bytes.Index(testMessage, []byte("\n\n"))
	_, err = MessageBody(testMessage[:high])
	if _, ok := err.(BodyNotFoundError); !ok {
		t.Errorf("MessageBody(…) have returned %T insted of %T", err, BodyNotFoundError{})
	}
}

func TestSplitFrom(t *testing.T) {
	tests := []struct {
		value []byte
		name  []byte
		email []byte
	}{
		{[]byte("author@example.com"), nil, []byte("author@example.com")},
		{[]byte("<author@example.com>"), nil, []byte("author@example.com")},
		{[]byte("Author <author@example.com>"), []byte("Author"), []byte("author@example.com")},
	}

	for _, tt := range tests {
		name, email := SplitFromTo(tt.value)
		if bytes.Compare(name, tt.name) != 0 || bytes.Compare(email, tt.email) != 0 {
			t.Errorf("SplitFrom(%q) have returned (%q, %q) instead of (%q, %q)", tt.value, name, email, tt.name, tt.email)
		}
	}
}

func TestSplitMessage(t *testing.T) {
	fromLineWant := []byte("MAILER-DAEMON Fri Jul  8 12:08:34 2011")
	headersWant := map[string][]byte{
		"From":    []byte("Author <author@example.com>"),
		"To":      []byte("Recipient <recipient@example.com>"),
		"Subject": []byte("Sample message 1"),
	}
	bodyWant := []byte("This is the body.\n>From (should be escaped).\nThere are 3 lines.")

	fromLine, headers, body, err := SplitMessage(testMessage)
	if err != nil {
		t.Errorf("SplitMessage(…) should't have returned %T (%[1]s)", err)
	} else {
		if bytes.Compare(fromLine, fromLineWant) != 0 {
			t.Errorf("SplitMessage(…) have returned fromLine=%q instead of fromLine=%q", fromLine, fromLineWant)
		}
		if !reflect.DeepEqual(headers, headersWant) {
			t.Errorf("SplitMessage(…) have returned\nheaders = %s instead of\nheaders = %s", headers, headersWant)
		}
		if bytes.Compare(body, bodyWant) != 0 {
			t.Errorf("SplitMessage(…) have returned\nbody = %#v instead of\nbody = %#v", body, bodyWant)
		}
	}
}

func TestTrimPatch(t *testing.T) {
	tests := []struct {
		value        []byte
		trimmedValue []byte
	}{
		{[]byte("Sample message 1"), []byte("Sample message 1")},
		{[]byte("[PATCH] Sample message 1"), []byte("Sample message 1")},
		{[]byte("[PATCH 1/1] Sample message 1"), []byte("Sample message 1")},
	}

	for _, tt := range tests {
		trimmedValue := TrimPatchPrefix(tt.value)
		if bytes.Compare(trimmedValue, tt.trimmedValue) != 0 {
			t.Errorf("TrimSubject(%q) have returned %q instead of %q", tt.value, trimmedValue, tt.trimmedValue)
		}
	}
}
