/*
package asaka provides consumer facility to read monitor data from Asaka
*/
package asaka

import (
	"errors"
	"strings"

	"github.com/hpcloud/tail"
	"github.com/ksang/hana/datasource"
	"github.com/olebedev/config"
)

type asaka struct {
	filePath string
	quitCh   chan struct{}
	running  bool
}

func New(conf string) (datasource.Consumer, error) {
	cfg, err := config.ParseYaml(conf)
	if err != nil {
		return nil, err
	}
	fp, err := cfg.String("filepath")
	if err != nil {
		return nil, err
	}
	return &asaka{
		filePath: fp,
		quitCh:   make(chan struct{}, 1),
	}, nil
}

func (a *asaka) Start() (chan string, error) {
	t, err := tail.TailFile(a.filePath, tail.Config{Follow: true})
	if err != nil {
		return nil, err
	}
	ret := make(chan string, 1)
	go func() {
		for {
			select {
			case <-a.quitCh:
				t.Stop()
				a.running = false
				close(ret)
				return
			case line := <-t.Lines:
				if line != nil {
					ret <- strings.TrimSuffix(line.Text, "\r")
				}
			}
		}
	}()
	a.running = true
	return ret, nil
}

func (a *asaka) Stop() error {
	if !a.running {
		return errors.New("not running")
	}
	a.quitCh <- struct{}{}
	return nil
}
