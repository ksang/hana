package asaka

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	testLogFile = "test.log"
)

func writeLogFile(t *testing.T) {
	for i := 0; i < 20; i++ {
		data := fmt.Sprintf("%d,%d\n", i, time.Now().Unix())
		fl, err := os.OpenFile(testLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			t.Fatal(err)
			return
		}
		_, err = fl.WriteString(data)
		if err != nil {
			t.Fatal(err)
			return
		}
		fl.Close()
		time.Sleep(time.Second)
	}
}

func TestAsakaConsumer(t *testing.T) {
	cons, err := New(fmt.Sprintf("datasource:\n  asaka\nfilepath:\n  %s", testLogFile))
	if err != nil {
		t.Error(err)
	}
	out, err := cons.Start()
	if err != nil {
		t.Error(err)
	}
	go func() {
		for line := range out {
			t.Log(line)
		}
	}()
	go writeLogFile(t)
	time.Sleep(time.Second * 10)
	if err := cons.Stop(); err != nil {
		t.Error(err)
	}
	os.Remove(testLogFile)
}
