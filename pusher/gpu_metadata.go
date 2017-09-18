package pusher

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/olebedev/config"
	"github.com/prometheus/client_golang/prometheus"
)

type GPUMetaLogType int

const (
	_ GPUMetaLogType = iota
	GPU_UTIL
	GPU_MEMORY
	GPU_TEMPERATURE
	PCIE_BW_RX
	PCIE_BW_TX
)

type gpu_meta struct {
	pushUrl string
	source  chan string
	quitCh  chan struct{}
	running bool
}

var (
	gpuLabelList  = []string{"id", "name"}
	pcieLabelList = []string{"id", "name"}
	gpuUtilMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gpu_utilization",
			Help: "gpu core utlization",
		},
		gpuLabelList,
	)
	gpuMemMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gpu_memory_utilization",
			Help: "gpu memory utlization",
		},
		gpuLabelList,
	)
	gpuTempMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gpu_temperature",
			Help: "gpu temperature in C degree",
		},
		gpuLabelList,
	)
	pcieBWRXMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pcie_bandwidth_rx",
			Help: "pcie bandwidth rx in MB",
		},
		pcieLabelList,
	)
	pcieBWTXMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pcie_bandwidth_tx",
			Help: "pcie bandwidth tx in MB",
		},
		pcieLabelList,
	)
)

func NewGPUMeta(conf string) (Pusher, error) {
	cfg, err := config.ParseYaml(conf)
	if err != nil {
		return nil, err
	}
	pushurl, err := cfg.String("pushurl")
	if err != nil {
		pushurl = ""
	}
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(gpuUtilMetric)
	prometheus.MustRegister(gpuMemMetric)
	prometheus.MustRegister(gpuTempMetric)

	return &gpu_meta{
		pushUrl: pushurl,
		quitCh:  make(chan struct{}, 1),
	}, nil
}

func (g *gpu_meta) Start(src chan string) error {
	g.source = src
	go func() {
		for {
			select {
			case <-g.quitCh:
				g.running = false
				return
			case line := <-g.source:
				g.ParseAndPush(line)
			}
		}
	}()
	return nil
}

func (g *gpu_meta) Stop() error {
	if !g.running {
		return errors.New("not running")
	}
	g.quitCh <- struct{}{}
	return nil
}

func (g *gpu_meta) ParseAndPush(data string) {
	dataList := strings.Split(data, ",")
	if len(dataList) < 2 {
		// ignore
		return
	}
	logType, err := strconv.ParseInt(dataList[1], 10, 8)
	if err != nil {
		log.Println("failed to parse gpu_meta log type,", err)
		return
	}
	if len(dataList) < 5 {
		log.Println("incorrect gpu_meta log format")
		return
	}

	gpu_id := dataList[2]
	gpu_name := dataList[3]

	value, err := strconv.ParseFloat(strings.TrimSpace(dataList[4]), 64)
	if err != nil {
		log.Println("data format error for parsing", err)
		return
	}

	labels := prometheus.Labels{
		"id":   gpu_id,
		"name": gpu_name,
	}
	if len(g.pushUrl) == 0 {
		log.Printf("data parsed: TYPE: %d GPUID: %s NAME: %s VALUE: %f",
			logType, gpu_id, gpu_name, value)
		return
	}

	switch GPUMetaLogType(logType) {
	case GPU_UTIL:
		gpuUtilMetric.With(labels).Set(value)
	case GPU_MEMORY:
		gpuUtilMetric.With(labels).Set(value)
	case GPU_TEMPERATURE:
		gpuUtilMetric.With(labels).Set(value)
	case PCIE_BW_RX:
		pcieBWRXMetric.With(labels).Set(value)
	case PCIE_BW_TX:
		pcieBWTXMetric.With(labels).Set(value)
	default:
		log.Println("unknown gpu meta log type,", logType)
	}
	return
}
