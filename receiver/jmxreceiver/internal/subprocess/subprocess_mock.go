package subprocess

import (
	"bufio"
	"context"
	"os"

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
	return nil
}

func (subprocess *SubprocessMock) Start(context.Context) error {
	reader, writer, _ := os.Pipe()
	writer.Write([]byte(subprocess.errorStr))
	collectStdout(bufio.NewScanner(reader), subprocess.Stdout(), subprocess.logger)
	return nil
}
