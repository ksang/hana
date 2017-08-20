package pusher

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/olebedev/config"
	"github.com/prometheus/client_golang/prometheus"
)

type asaka struct {
	pushUrl    string
	source     chan string
	quitCh     chan struct{}
	apiNameMap map[string]interface{}
	running    bool
}

var (
	labelList    = []string{"session", "client_id", "api"}
	runtimeMetrc = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asaka_api_running_time",
			Help: "api total running time",
		},
		labelList,
	)
	callcountMetrc = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asaka_api_call_count",
			Help: "api total call count",
		},
		labelList,
	)
	totalsizeMetrc = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asaka_api_total_size",
			Help: "api total size",
		},
		labelList,
	)
)

func NewAsaka(conf string) (Pusher, error) {
	cfg, err := config.ParseYaml(conf)
	if err != nil {
		return nil, err
	}
	apimap, err := cfg.Map("apinamemap")
	if err != nil {
		return nil, err
	}
	pushurl, err := cfg.String("pushurl")
	if err != nil {
		pushurl = ""
	}
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(runtimeMetrc)
	prometheus.MustRegister(callcountMetrc)
	prometheus.MustRegister(totalsizeMetrc)

	return &asaka{
		pushUrl:    pushurl,
		quitCh:     make(chan struct{}, 1),
		apiNameMap: apimap,
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
	if len(dataList) < 7 {
		// ignore
		return
	}
	sessid := dataList[1]
	clientid := dataList[2]
	apiname, ok := a.apiNameMap[dataList[3]]
	if !ok {
		apiname = dataList[3]
	}
	runtime, err := strconv.ParseUint(dataList[4], 10, 64)
	if err != nil {
		log.Println("Data format error for parsing running time,", err)
		return
	}
	callcount, err := strconv.ParseUint(dataList[5], 10, 64)
	if err != nil {
		log.Println("Data format error for parsing calling count,", err)
		return
	}
	size, err := strconv.ParseUint(dataList[6], 10, 64)
	if err != nil {
		log.Println("Data format error for parsing size,", err)
		return
	}
	if len(a.pushUrl) == 0 {
		log.Printf("Data parsed: SESS: %s CLIENT_ID: %s API_NAME: %s RUNTIME: %d CALLCOUNT: %d SIZE: %d",
			sessid, clientid, apiname, runtime, callcount, size)
		return
	}
	labels := prometheus.Labels{
		"session":   sessid,
		"client_id": clientid,
		"api":       apiname.(string),
	}

	runtimeMetrc.With(labels).Set(float64(runtime))
	callcountMetrc.With(labels).Set(float64(callcount))
	totalsizeMetrc.With(labels).Set(float64(size))

	return
}
