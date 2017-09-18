package pusher

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/olebedev/config"
	"github.com/prometheus/client_golang/prometheus"
)

type AsakaLogType int

const (
	_ AsakaLogType = iota
	MONITOR_API
	MONITOR_KERNEL
)

type asaka struct {
	pushUrl string
	source  chan string
	quitCh  chan struct{}
	running bool
}

var (
	apiLabelList     = []string{"session", "client_id", "api"}
	apiRuntimeMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asaka_api_running_time",
			Help: "api total running time",
		},
		apiLabelList,
	)
	apiCallcountMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asaka_api_call_count",
			Help: "api total call count",
		},
		apiLabelList,
	)
	apiTotalsizeMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asaka_api_total_size",
			Help: "api total size",
		},
		apiLabelList,
	)

	kernelLabelList     = []string{"session", "client_id", "name"}
	kernelRuntimeMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asaka_kernel_running_time",
			Help: "kernel total running time",
		},
		kernelLabelList,
	)
	kernelCallcountMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asaka_kernel_call_count",
			Help: "kernel total call count",
		},
		kernelLabelList,
	)
	kernelBlocknumMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asaka_kernel_block_num",
			Help: "kernel total block num",
		},
		kernelLabelList,
	)
	kernelThreadnumMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asaka_kernel_thread_num",
			Help: "kernel total thread num",
		},
		kernelLabelList,
	)
)

func NewAsaka(conf string) (Pusher, error) {
	cfg, err := config.ParseYaml(conf)
	if err != nil {
		return nil, err
	}
	pushurl, err := cfg.String("pushurl")
	if err != nil {
		pushurl = ""
	}
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(apiRuntimeMetric)
	prometheus.MustRegister(apiCallcountMetric)
	prometheus.MustRegister(apiTotalsizeMetric)
	prometheus.MustRegister(kernelRuntimeMetric)
	prometheus.MustRegister(kernelCallcountMetric)
	prometheus.MustRegister(kernelBlocknumMetric)
	prometheus.MustRegister(kernelThreadnumMetric)

	return &asaka{
		pushUrl: pushurl,
		quitCh:  make(chan struct{}, 1),
	}, nil
}

func (a *asaka) Start(src chan string) error {
	a.source = src
	go func() {
		for {
			select {
			case <-a.quitCh:
				a.running = false
				return
			case line := <-a.source:
				a.ParseAndPush(line)
			}
		}
	}()
	return nil
}

func (a *asaka) Stop() error {
	if !a.running {
		return errors.New("not running")
	}
	a.quitCh <- struct{}{}
	return nil
}

func (a *asaka) ParseAndPush(data string) {
	dataList := strings.Split(data, ",")
	if len(dataList) < 2 {
		// ignore
		return
	}
	logType, err := strconv.ParseInt(dataList[1], 10, 8)
	if err != nil {
		log.Println("failed to parse asaka log type,", err)
		return
	}
	switch AsakaLogType(logType) {
	case MONITOR_API:
		a.parseAndPushAPI(dataList)
	case MONITOR_KERNEL:
		a.parseAndPushKernel(dataList)
	default:
		log.Println("unknown asaka log type,", logType)
	}
	return
}

func (a *asaka) parseAndPushAPI(dataList []string) {
	sessid := dataList[2]
	clientid := dataList[3]
	apiname := dataList[4]
	runtime, err := strconv.ParseUint(dataList[5], 10, 64)
	if err != nil {
		log.Println("data format error for parsing running time,", err)
		return
	}
	callcount, err := strconv.ParseUint(dataList[6], 10, 64)
	if err != nil {
		log.Println("data format error for parsing calling count,", err)
		return
	}
	size, err := strconv.ParseUint(dataList[7], 10, 64)
	if err != nil {
		log.Println("data format error for parsing size,", err)
		return
	}
	if len(a.pushUrl) == 0 {
		log.Printf("data parsed: SESS: %s CLIENT_ID: %s API_NAME: %s RUNTIME: %d CALLCOUNT: %d SIZE: %d",
			sessid, clientid, apiname, runtime, callcount, size)
		return
	}
	labels := prometheus.Labels{
		"session":   sessid,
		"client_id": clientid,
		"api":       apiname,
	}

	apiRuntimeMetric.With(labels).Set(float64(runtime))
	apiCallcountMetric.With(labels).Set(float64(callcount))
	apiTotalsizeMetric.With(labels).Set(float64(size))
}

func (a *asaka) parseAndPushKernel(dataList []string) {
	sessid := dataList[2]
	clientid := dataList[3]
	kernelname := dataList[5]
	runtime, err := strconv.ParseUint(dataList[6], 10, 64)
	if err != nil {
		log.Println("data format error for parsing running time,", err)
		return
	}
	callcount, err := strconv.ParseUint(dataList[7], 10, 64)
	if err != nil {
		log.Println("data format error for parsing calling count,", err)
		return
	}
	blocknum, err := strconv.ParseUint(dataList[8], 10, 64)
	if err != nil {
		log.Println("data format error for parsing blocknum,", err)
		return
	}
	threadnum, err := strconv.ParseUint(dataList[9], 10, 64)
	if err != nil {
		log.Println("data format error for parsing threadnum,", err)
		return
	}
	if len(a.pushUrl) == 0 {
		log.Printf("data parsed: SESS: %s CLIENT_ID: %s KERNEL_NAME: %s RUNTIME: %d CALLCOUNT: %d BLOCK_NUM: %d THREAD_NUM: %d",
			sessid, clientid, kernelname, runtime, callcount, blocknum, threadnum)
		return
	}
	labels := prometheus.Labels{
		"session":   sessid,
		"client_id": clientid,
		"name":      kernelname,
	}

	kernelRuntimeMetric.With(labels).Set(float64(runtime))
	kernelCallcountMetric.With(labels).Set(float64(callcount))
	kernelBlocknumMetric.With(labels).Set(float64(blocknum))
	kernelThreadnumMetric.With(labels).Set(float64(threadnum))
}
