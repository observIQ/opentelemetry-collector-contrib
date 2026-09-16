// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package windows

import (
	"context"
	"encoding/hex"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/require"
)

// canaryValue is 13 bytes, so every clone lands in a 16-byte tiny-allocator block: the same size class
// as the short strings (namespaces, log types, attribute keys) that a freed uint32 out-parameter ends
// up sharing a block with.
const canaryValue = "canary-string"

// addrToPointer converts an integer address back to a pointer, standing in for the kernel side of a
// Win32 out-parameter write.
func addrToPointer(addr uintptr) unsafe.Pointer { return *(*unsafe.Pointer)(unsafe.Pointer(&addr)) }

// staleWriteProc stands in for EvtGetLogInfo as the kernel sees it: it receives the address of the
// caller's BufferUsed as a plain integer, the Go heap keeps moving while the call is in flight, and it
// finally stores BufferUsed = sizeof(EVT_VARIANT) = 16 through that address.
type staleWriteProc struct {
	mu     sync.Mutex
	keep   []string
	calls  atomic.Int64
	churns int
}

func (p *staleWriteProc) Call(a ...uintptr) (uintptr, uintptr, error) {
	bufferUsed := a[4] // fifth argument of EvtGetLogInfo, see evtGetLogInfo in api.go
	runtime.GC()
	runtime.GC()
	p.mu.Lock()
	for i := 0; i < p.churns; i++ {
		p.keep = append(p.keep, strings.Clone(canaryValue))
	}
	p.mu.Unlock()
	*(*uint32)(addrToPointer(bufferUsed)) = 16
	p.calls.Add(1)
	return 1, 0, nil
}

func (p *staleWriteProc) corrupted() (int, []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var samples []string
	n := 0
	for _, s := range p.keep {
		if s != canaryValue {
			n++
			if len(samples) < 3 {
				samples = append(samples, hex.EncodeToString([]byte(s)))
			}
		}
	}
	return n, samples
}

// TestPollAndReadKeepsEvtGetLogInfoOutParamsAlive drives the production pollAndRead loop with the
// channel-size sample enabled. The BufferUsed out-parameter handed to EvtGetLogInfo must stay valid
// until the call returns; if it is collected mid-call, the API's write lands in whatever reused the
// memory.
func TestPollAndReadKeepsEvtGetLogInfoOutParamsAlive(t *testing.T) {
	orig := getLogInfoProc
	t.Cleanup(func() { getLogInfoProc = orig })
	proc := &staleWriteProc{churns: 20000}
	getLogInfoProc = proc

	input := newTestInputWithTelemetry(&mockTelemetry{})
	input.channel = "Application"
	input.logHandle = 1 // enables the EvtGetLogInfo sample; readBatch itself fails fast without a subscription
	input.pollInterval = time.Millisecond
	input.maxReads = 1

	ctx, cancel := context.WithCancel(t.Context())
	input.wg.Add(1)
	go input.pollAndRead(ctx)
	require.Eventually(t, func() bool { return proc.calls.Load() >= 200 }, 2*time.Minute, 10*time.Millisecond)
	cancel()
	input.wg.Wait()

	n, samples := proc.corrupted()
	t.Logf("EvtGetLogInfo calls=%d canaries=%d corrupted=%d samples=%v", proc.calls.Load(), len(proc.keep), n, samples)
	require.Zero(t, n, "EvtGetLogInfo's BufferUsed pointer was freed before the call completed; the write corrupted %d live strings, e.g. %v", n, samples)
}
