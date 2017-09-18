package pusher

import (
	"testing"
	"time"
)

var (
	gpu_monitor_data = []string{
		"2017/09/18 00:28:08.188,2,1,Tesla P100-SXM2-16GB,0",
		"2017/09/18 00:28:08.188,3,1,Tesla P100-SXM2-16GB,27",
		"2017/09/18 00:28:08.188,4,1,Tesla P100-SXM2-16GB,0",
		"2017/09/18 00:28:08.188,5,1,Tesla P100-SXM2-16GB,0",
		"2017/09/18 00:28:08.189,1,2,Tesla P100-SXM2-16GB,0",
		"2017/09/18 00:28:08.189,2,2,Tesla P100-SXM2-16GB,0",
		"2017/09/18 00:28:08.189,3,2,Tesla P100-SXM2-16GB,25",
		"2017/09/18 00:28:08.189,4,2,Tesla P100-SXM2-16GB,0",
		"2017/09/18 00:28:08.189,5,2,Tesla P100-SXM2-16GB,0",
		"2017/09/18 00:28:08.190,1,3,Tesla P100-SXM2-16GB,0"}
)

func TestGpuMetaPusher(t *testing.T) {
	pusher, err := NewGPUMeta("")
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
		for _, data := range gpu_monitor_data {
			time.Sleep(200 * time.Millisecond)
			src <- data
		}
	}()
	time.Sleep(5 * time.Second)
	pusher.Stop()
}
