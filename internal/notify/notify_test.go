package notify_test

import (
	"testing"

	"github.com/theverything/communique/internal/notify"
)

func TestNotify(t *testing.T) {
	client := notify.New()

	i, err := client.Write([]byte("foo"))
	if err != nil {
		t.Error("Write should not error")
	}

	if i != 3 {
		t.Error("i should equal len(payload)")
	}

	o := <-client.C

	if string(o) != "foo" {
		t.Errorf(`Incorrect output expected "%s" received "%s"`, "foo", string(o))
	}
}
