package subprocess

import (
	"bufio"
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
)

type SubprocessMock struct {
	stdout   chan string
	logger   *zap.Logger
	errorStr string
}

func NewMockSubprocess(logger *zap.Logger, errorStr string) *SubprocessMock {
	return &SubprocessMock{
		stdout:   make(chan string),
		logger:   logger,
		errorStr: errorStr,
	}
}

func (subprocess *SubprocessMock) Stdout() chan string {
	return subprocess.stdout
}

func (subprocess *SubprocessMock) Shutdown(context.Context) error {
	time.Sleep(time.Second)
	close(subprocess.stdout)
	return nil
}

func (subprocess *SubprocessMock) Start(context.Context) error {
	go collectStdout(bufio.NewScanner(strings.NewReader(subprocess.errorStr)), subprocess.Stdout(), subprocess.logger)
	return nil
}
