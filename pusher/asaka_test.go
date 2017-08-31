package pusher

import (
	"testing"
	"time"
)

var (
	asaka_monitor_data = []string{
		"1502970051,1,0,2,cuda_init,2,1,0",
		"1502970051,1,0,2,cuda_init,221,1,33554432",
		"1502970058,1,0,3,cuda_malloc,13,1,0",
		"1502970058,1,0,3,cuModuleLoadData,2,1,0",
		"1504171516,1,0,1,cuModuleLoadData,379,1,0",
		"1504171516,1,0,1,cuModuleLoadData,4,1,0",
		"1504171516,1,0,1,TEST,983,4,2097154",
		"1504171516,2,0,1,0x7fb7ec062910,_Z13FFT512_deviceI6float2fEvPT_,130,10,2560,640",
		"1504171516,2,0,1,0x7fb7ec05b500,_Z14IFFT512_deviceI6float2fEvPT_,106,10,2560,640",
		"1504171516,2,0,1,0x7fb7ec05fcb0,_Z13chk512_deviceI6float2EvPKT_iPc,104,10,1280,640",
		"1504171516,2,0,1,0x7fb7ec06e680,_Z13FFT512_deviceI7double2dEvPT_,109,10,1280,640",
		"1504171516,2,0,1,0x7fb7ec066fb0,_Z14IFFT512_deviceI7double2dEvPT_,97,10,1280,640",
		"1504171516,2,0,1,0x7fb7ec06b9a0,_Z13chk512_deviceI7double2EvPKT_iPc,88,10,640,64",
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
			time.Sleep(200 * time.Millisecond)
			src <- data
		}
	}()
	time.Sleep(5 * time.Second)
	pusher.Stop()
}
