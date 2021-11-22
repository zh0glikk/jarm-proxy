package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/google/jsonapi"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/jarm-proxy/internal/config"
)

type Request struct {
	Target string
	Port   string
}

type requestsSender struct {
	logger *logan.Entry
	cfg    config.Config

	targets    chan Request
	concurency int
}

func NewRequestsSender(cfg config.Config) *requestsSender {
	return &requestsSender{
		logger:     cfg.Log(),
		cfg:        cfg,
		targets:    make(chan Request),
		concurency: 100,
	}
}

func (s *requestsSender) Run(ctx context.Context) error {
	go s.reader()

	var wg sync.WaitGroup
	wg.Add(s.concurency)

	for i := 0; i < s.concurency; i++ {
		go s.worker(&wg)
	}

	wg.Wait()
	return nil
}

func (s *requestsSender) reader() error {
	defer close(s.targets)
	file, err := os.Open("tmp3.csv")
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer file.Close()

	csvReader := csv.NewReader(file)

	for {
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		s.targets <- Request{
			Target: record[0],
			Port:   record[1],
		}
	}

	return nil
}

func (s *requestsSender) process(target Request) error {
	cmd := exec.Command("curl", "-x", "http://127.0.0.1:8090", fmt.Sprintf("%s:%s", target.Target, target.Port))
	out, err := cmd.Output()
	if err != nil {
		s.logger.WithError(err).Error("failed to execute cmd")
		return errors.Wrap(err, "failed to execute cmd")
	}

	var tmp jsonapi.ErrorObject

	err = json.Unmarshal(out, &tmp)
	if err != nil {
		return nil
	}

	if tmp.Status == fmt.Sprintf("%d", http.StatusForbidden) {
		s.logger.WithFields(logan.F{
			"fingerprint": tmp.Detail,
			"target":      target.Target,
			"port":        target.Port,
		}).Warn("malicious target")
	}

	return nil
}

func (s *requestsSender) worker(wg *sync.WaitGroup) {
	defer wg.Done()

	for value := range s.targets {
		err := s.process(value)
		if err != nil {
			s.logger.WithError(err).Error("failed to process request")
		}
	}
}
