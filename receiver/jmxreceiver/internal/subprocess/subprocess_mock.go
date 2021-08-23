package subprocess

import (
	"bufio"
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Mock struct {
	stdout   chan string
	logger   *zap.Logger
	errorStr string
}

func NewMockSubprocess(logger *zap.Logger, errorStr string) *Mock {
	return &Mock{
		stdout:   make(chan string),
		logger:   logger,
		errorStr: errorStr,
	}
}

func (subprocess *Mock) Stdout() chan string {
	return subprocess.stdout
}

func (subprocess *Mock) Shutdown(context.Context) error {
	time.Sleep(time.Second)
	close(subprocess.stdout)
	return nil
}

func (subprocess *Mock) Start(context.Context) error {
	go collectStdout(bufio.NewScanner(strings.NewReader(subprocess.errorStr)), subprocess.Stdout(), subprocess.logger)
	return nil
}
