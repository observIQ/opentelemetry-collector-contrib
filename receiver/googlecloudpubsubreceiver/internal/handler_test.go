// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"testing"
	"time"

	pubsub "cloud.google.com/go/pubsub/v2/apiv1"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"cloud.google.com/go/pubsub/v2/pstest"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudpubsubreceiver/internal/metadata"
)

type flakySubscriberClient struct {
	SubscriberClient
	// failures is the number of upcoming StreamingPull calls that must fail
	failures atomic.Int32
	calls    atomic.Int32
}

func (c *flakySubscriberClient) StreamingPull(ctx context.Context, opts ...gax.CallOption) (pubsubpb.Subscriber_StreamingPullClient, error) {
	c.calls.Add(1)
	if c.failures.Load() > 0 {
		c.failures.Add(-1)
		return nil, status.Error(codes.Unauthenticated, "transport: per-RPC creds failed due to error: oauth2: cannot fetch token")
	}
	return c.SubscriberClient.StreamingPull(ctx, opts...)
}

func createServer(ctx context.Context, t *testing.T) (cleanupFn func(), srv *pstest.Server, client SubscriberClient) {
	srv = pstest.NewServer()

	var copts []option.ClientOption
	var dialOpts []grpc.DialOption
	conn, err := grpc.NewClient(srv.Addr, append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))...)
	assert.NoError(t, err)

	cleanupFn = func() {
		assert.NoError(t, srv.Close())
		assert.NoError(t, conn.Close())
	}

	copts = append(copts, option.WithGRPCConn(conn))
	_, err = srv.GServer.CreateTopic(ctx, &pubsubpb.Topic{
		Name: "projects/my-project/topics/otlp",
	})
	assert.NoError(t, err)
	_, err = srv.GServer.CreateSubscription(ctx, &pubsubpb.Subscription{
		Topic:              "projects/my-project/topics/otlp",
		Name:               "projects/my-project/subscriptions/otlp",
		AckDeadlineSeconds: 10,
	})
	assert.NoError(t, err)

	client, err = pubsub.NewSubscriptionAdminClient(ctx, copts...)
	assert.NoError(t, err)
	return cleanupFn, srv, client
}

func createHandlerWithClient(ctx context.Context, t *testing.T, client SubscriberClient, callback func(context.Context, *pubsubpb.ReceivedMessage) error) *StreamHandler {
	settings := receivertest.NewNopSettings(metadata.Type)
	telemetryBuilder, _ := metadata.NewTelemetryBuilder(settings.TelemetrySettings)
	handler, err := NewHandler(ctx, settings, telemetryBuilder, client, "client-id", "projects/my-project/subscriptions/otlp",
		nil, callback)
	assert.NoError(t, err)
	return handler
}

func createHandler(ctx context.Context, t *testing.T) (cleanupFn func(), srv *pstest.Server, handler *StreamHandler) {
	cleanupFn, srv, client := createServer(ctx, t)
	handler = createHandlerWithClient(ctx, t, client, func(context.Context, *pubsubpb.ReceivedMessage) error {
		return nil
	})
	return cleanupFn, srv, handler
}

func TestCancelStream(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	cleanupFn, srv, handler := createHandler(ctx, t)
	defer cleanupFn()

	srv.Publish("projects/my-project/topics/otlp", []byte{}, map[string]string{
		"ce-type":      "org.opentelemetry.otlp.traces.v1",
		"content-type": "application/protobuf",
	})
	handler.RecoverableStream(ctx)
	go func() {
		time.Sleep(100 * time.Millisecond)
		handler.CancelNow()
	}()
	handler.Wait()
}

// TestRecoverableStreamSurvivesFailedReinit makes the server close every stream shortly after it is
// created and makes StreamingPull fail twice on re-initialization. The handler must not spawn its
// goroutines against the nil stream and must recover once StreamingPull succeeds again.
func TestRecoverableStreamSurvivesFailedReinit(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	cleanupFn, srv, client := createServer(ctx, t)
	defer cleanupFn()
	// every stream is closed by the server shortly after creation, forcing the recovery loop
	srv.SetStreamTimeout(200 * time.Millisecond)

	flaky := &flakySubscriberClient{SubscriberClient: client}
	received := make(chan string, 16)
	handler := createHandlerWithClient(ctx, t, flaky, func(_ context.Context, message *pubsubpb.ReceivedMessage) error {
		received <- string(message.Message.Data)
		return nil
	})
	// the initial stream was created by NewHandler; the next two re-initializations fail
	flaky.failures.Store(2)
	handler.RecoverableStream(ctx)
	defer handler.CancelNow()

	// initial + 2 failed + 1 successful re-initialization
	require.Eventually(t, func() bool { return flaky.calls.Load() >= 4 }, 10*time.Second, 10*time.Millisecond)
	assert.Equal(t, int32(0), flaky.failures.Load())

	srv.Publish("projects/my-project/topics/otlp", []byte("after-recovery"), nil)
	select {
	case data := <-received:
		assert.Equal(t, "after-recovery", data)
	case <-time.After(10 * time.Second):
		t.Fatal("message was not received after stream recovery")
	}
}

// TestRecoverableStreamStopsWhenContextCanceled cancels the context the handler was started with,
// without calling CancelNow. The recovery loop can never recreate a stream with a canceled context,
// so it must stop instead of retrying (or crashing) forever.
func TestRecoverableStreamStopsWhenContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cleanupFn, srv, handler := createHandler(ctx, t)
	defer cleanupFn()

	srv.Publish("projects/my-project/topics/otlp", []byte{}, nil)
	handler.RecoverableStream(ctx)
	time.Sleep(100 * time.Millisecond)
	cancel()

	done := make(chan struct{})
	go func() {
		handler.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("recovery loop did not stop after its context was canceled")
	}
	handler.CancelNow()
}

func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		retry int
		max   time.Duration
	}{
		{
			retry: 0,
			max:   time.Duration(0),
		},
	}
	for i := 1; i <= 11; i++ {
		maxBackoff := min(time.Duration(250.0*math.Pow(2, float64(i-1)))*time.Millisecond, time.Duration(2)*time.Minute)
		tests = append(tests, struct {
			retry int
			max   time.Duration
		}{
			retry: i,
			max:   maxBackoff,
		})
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("retry-%d", tt.retry), func(t *testing.T) {
			for range 10 {
				backoff := exponentialBackoff(tt.retry)
				minBackoffDueToJitter := time.Duration(0.7*float64(tt.max.Milliseconds())) * time.Millisecond
				assert.Condition(t, func() bool { return backoff <= tt.max },
					"exponentialBackoff %s should not go over max %s", backoff.String(), tt.max.String())
				assert.Condition(t, func() bool { return backoff >= minBackoffDueToJitter },
					"exponentialBackoff %s should not go under min (due to jitter) %s", backoff.String(), minBackoffDueToJitter.String())
			}
		})
	}
}
