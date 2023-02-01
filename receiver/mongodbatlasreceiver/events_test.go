package mongodbatlasreceiver

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/atlas/mongodbatlas"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/comparetest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/comparetest/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver/internal"
)

func TestStartAndShutdown(t *testing.T) {
	cases := []struct {
		desc                string
		getConfig           func() *Config
		expectedStartErr    error
		expectedShutdownErr error
	}{
		{
			desc: "valid config",
			getConfig: func() *Config {
				cfg := createDefaultConfig().(*Config)
				cfg.Events = &EventsConfig{
					Projects: []*ProjectConfig{
						{
							Name: testProjectName,
						},
					},
					PollInterval: time.Minute,
				}
				return cfg
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sink := &consumertest.LogsSink{}
			r := newEventsReceiver(receivertest.NewNopCreateSettings(), tc.getConfig(), sink)
			err := r.Start(context.Background(), componenttest.NewNopHost())
			if tc.expectedStartErr != nil {
				require.ErrorContains(t, err, tc.expectedStartErr.Error())
				return
			}
			require.NoError(t, err)
			err = r.Shutdown(context.Background())
			if tc.expectedShutdownErr != nil {
				require.ErrorContains(t, err, tc.expectedShutdownErr.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestContextDone(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Events = &EventsConfig{
		Projects: []*ProjectConfig{
			{
				Name: testProjectName,
			},
		},
		PollInterval: 500 * time.Millisecond,
	}
	sink := &consumertest.LogsSink{}
	r := newEventsReceiver(receivertest.NewNopCreateSettings(), cfg, sink)
	mClient := &mockEventsClient{}
	mClient.setupMock(t)
	r.client = mClient

	ctx, cancel := context.WithCancel(context.Background())
	err := r.Start(ctx, componenttest.NewNopHost())
	require.NoError(t, err)
	cancel()

	require.Never(t, func() bool {
		return sink.LogRecordCount() > 0
	}, 500*time.Millisecond, 2*time.Second)
}

func TestPoll(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Events = &EventsConfig{
		Projects: []*ProjectConfig{
			{
				Name: testProjectName,
			},
		},
		PollInterval: time.Second,
	}

	sink := &consumertest.LogsSink{}
	r := newEventsReceiver(receivertest.NewNopCreateSettings(), cfg, sink)
	mClient := &mockEventsClient{}
	mClient.setupMock(t)
	r.client = mClient

	err := r.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return sink.LogRecordCount() > 0
	}, 5*time.Second, 1*time.Second)

	expected, err := golden.ReadLogs(filepath.Join("testdata", "events", "golden", "events.json"))
	require.NoError(t, err)

	logs := sink.AllLogs()[0]
	require.NoError(t, comparetest.CompareLogs(expected, logs, comparetest.IgnoreObservedTimestamp()))
}

type mockEventsClient struct {
	mock.Mock
}

func (mec *mockEventsClient) setupMock(t *testing.T) {
	mec.setupGetProject()
	mec.On("GetEvents", mock.Anything, mock.Anything, mock.Anything).Return(mec.loadTestEvents(t), false, nil)
}

func (mec *mockEventsClient) setupGetProject() {
	mec.On("GetProject", mock.Anything, mock.Anything).Return(&mongodbatlas.Project{
		ID:    testProjectID,
		OrgID: testOrgID,
		Name:  testProjectName,
		Links: []*mongodbatlas.Link{},
	}, nil)
}

func (mec *mockEventsClient) loadTestEvents(t *testing.T) []*mongodbatlas.Event {
	testEvents := filepath.Join("testdata", "events", "sample-payloads", "events.json")
	eventBytes, err := os.ReadFile(testEvents)
	require.NoError(t, err)

	var events []*mongodbatlas.Event
	err = json.Unmarshal(eventBytes, &events)
	require.NoError(t, err)
	return events
}

func (mec *mockEventsClient) GetProject(ctx context.Context, pID string) (*mongodbatlas.Project, error) {
	args := mec.Called(ctx, pID)
	return args.Get(0).(*mongodbatlas.Project), args.Error(1)
}

func (mec *mockEventsClient) GetEvents(ctx context.Context, pID string, opts *internal.GetEventsOptions) ([]*mongodbatlas.Event, bool, error) {
	args := mec.Called(ctx, pID, opts)
	return args.Get(0).([]*mongodbatlas.Event), args.Bool(1), args.Error(2)
}
