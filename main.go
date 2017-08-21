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
	UNKNOWN
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "d", "hana.conf", "configuration file location")
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
	default:
		return UNKNOWN
	}
}

func main() {
	flag.Parse()
	cfg, err := ioutil.ReadFile(configFile)
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
	default:
		return
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
