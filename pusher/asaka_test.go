package pusher

import (
	"testing"
	"time"
)

var (
	asaka_monitor_data = []string{
		"1502970051,0,2,1,2,1,0",
		"1502970051,0,2,5,221,1,33554432",
		"1502970058,0,3,0,13,1,0",
		"1502970058,0,3,1,2,1,0",
	}
	asaka_conf = "apinamemap:\n  \"1\": API_1\n  \"5\": API_5\n"
)

func TestAsakaPusher(t *testing.T) {
	pusher, err := NewAsaka(asaka_conf)
	if err != nil {
		t.Error(err)
		return
	}
	src := make(chan string, 1)
	err = pusher.Start(src)
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		for _, data := range asaka_monitor_data {
			time.Sleep(time.Second)
			src <- data
		}
	}()
	time.Sleep(5 * time.Second)
	pusher.Stop()
}
