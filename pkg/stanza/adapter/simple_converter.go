package adapter

import (
	"sync"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"go.opentelemetry.io/collector/pdata/plog"
)

type SimpleConverter struct {
	inputChan chan []*entry.Entry

	workerCount int

	outChan chan plog.Logs

	doneChan chan struct{}

	wg sync.WaitGroup
}

func NewSimpleConverter(inputChan chan []*entry.Entry, workerCount int, maxBatchSize int) *SimpleConverter {
	workers := workerCount
	if workerCount <= 0 {
		workers = 4
	}

	return &SimpleConverter{
		inputChan:   inputChan,
		workerCount: workers,
		outChan:     make(chan plog.Logs, workerCount*maxBatchSize),
		doneChan:    make(chan struct{}),
	}
}

func (s *SimpleConverter) Start() {
	s.wg.Add(s.workerCount)

	for i := 0; i < s.workerCount; i++ {
		go s.workerLoop()
	}
}

func (s *SimpleConverter) OutChannel() <-chan plog.Logs {
	return s.outChan
}

func (s *SimpleConverter) Stop() {
	close(s.doneChan)
	s.wg.Wait()
	close(s.outChan)
}

func (s *SimpleConverter) workerLoop() {
	defer s.wg.Done()

	resourceIDToLogs := make(map[uint64]plog.Logs)
	for {
		select {
		case <-s.doneChan:
			return
		case entries, ok := <-s.inputChan:
			if !ok {
				return
			}

			for _, e := range entries {
				logs := convert(e)
				resourceID := HashResource(e.Resource)

				// Aggregate
				plogs, ok := resourceIDToLogs[resourceID]
				if ok {
					lr := plogs.ResourceLogs().
						At(0).ScopeLogs().
						At(0).LogRecords().AppendEmpty()
					logs.CopyTo(lr)
					continue
				}

				plogs = plog.NewLogs()
				rls := plogs.ResourceLogs().AppendEmpty()

				resource := rls.Resource()
				insertToAttributeMap(e.Resource, resource.Attributes())

				lr := rls.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
				logs.CopyTo(lr)

				resourceIDToLogs[resourceID] = plogs
			}

			for r, plogs := range resourceIDToLogs {
				select {
				case <-s.doneChan:
					return
				case s.outChan <- plogs:
					delete(resourceIDToLogs, r)
				}
			}
		}
	}
}
