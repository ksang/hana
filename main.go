package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ksang/hana/datasource/asaka"
	"github.com/ksang/hana/pusher"
	"github.com/olebedev/config"
	"github.com/prometheus/client_golang/prometheus"
)

type DataSourceType int

const (
	_ DataSourceType = iota
	ASAKA
	GPUMETA
	UNKNOWN
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "d", "hana.conf", "configuration file location, use comma if you have multiple config files, listen address should be defined in first config file.")
}

func ParseDataSource(s string) DataSourceType {
	cfg, err := config.ParseYaml(s)
	if err != nil {
		log.Fatal("Failed to parse yaml config, err:", err)
	}
	ds, err := cfg.String("datasource")
	if err != nil {
		log.Fatal("Failed to get datasource type from config", err)
	}
	switch strings.ToLower(ds) {
	case "asaka":
		return ASAKA
	case "gpumeta":
		return GPUMETA
	default:
		return UNKNOWN
	}
}

func main() {
	flag.Parse()
	confFileList := strings.Split(configFile, ",")
	var conf string
	for _, confFile := range confFileList {
		cfg, err := ioutil.ReadFile(confFile)
		if err != nil {
			log.Fatal(err)
		}
		conf := string(cfg)
		ds := ParseDataSource(conf)
		switch ds {
		case ASAKA:
			consumer, err1 := asaka.New(conf)
			if err1 != nil {
				log.Fatal(err)
			}
			dataCh, err1 := consumer.Start()
			if err1 != nil {
				log.Fatal(err)
			}
			pusher, err1 := pusher.NewAsaka(conf)
			if err1 != nil {
				log.Fatal(err)
			}
			err1 = pusher.Start(dataCh)
			if err1 != nil {
				log.Fatal(err)
			}
		case GPUMETA:
			consumer, err1 := asaka.New(conf)
			if err1 != nil {
				log.Fatal(err)
			}
			dataCh, err1 := consumer.Start()
			if err1 != nil {
				log.Fatal(err)
			}
			pusher, err1 := pusher.NewGPUMeta(conf)
			if err1 != nil {
				log.Fatal(err)
			}
			err1 = pusher.Start(dataCh)
			if err1 != nil {
				log.Fatal(err)
			}
		default:
			log.Println("Unknown datasource type")
		}
	}
	ymalcfg, err := config.ParseYaml(conf)
	if err != nil {
		log.Fatal(err)
	}

	addr, err := ymalcfg.String("listenaddress")
	if err != nil {
		addr = ":9091"
	}
	http.Handle("/metrics", prometheus.Handler())
	go func() {
		log.Fatal(http.ListenAndServe(addr, nil))
	}()
	log.Println("Hana started at", addr+"/metrics")
	sigs := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- struct{}{}
	}()
	<-done
	fmt.Println("Signaled to terminate.")
}
