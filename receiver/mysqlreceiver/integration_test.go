package mysqlreceiver

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver/internal/metadata"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"
)

func mysqlContainer(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	var env = map[string]string{
		"MYSQL_DATABASE":      "otel",
		"MYSQL_USER":          "otel",
		"MYSQL_PASSWORD":      "otel",
		"MYSQL_ROOT_PASSWORD": "otel",
	}
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.mysql",
		},
		ExposedPorts: []string{"3306"},
		Env:          env,
	}

	require.NoError(t, req.Validate())

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	time.Sleep(time.Second * 6)
	return container
}

type MysqlIntegrationSuite struct {
	suite.Suite
}

func TestMysqlIntegration(t *testing.T) {
	suite.Run(t, new(MysqlIntegrationSuite))
}

func (suite *MysqlIntegrationSuite) TestHappyPath() {
	t := suite.T()
	container := mysqlContainer(t)
	defer container.Terminate(context.Background())

	sc := newMySQLScraper(zap.NewNop(), &Config{
		User:     "otel",
		Password: "otel",
		Endpoint: "127.0.0.1:3306",
	})
	err := sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)
	require.Equal(t, 1, rms.Len())

	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, 67, ms.Len())
	require.Equal(t, 14, len(metadata.M.Names()))

	metricsCount := map[string]int{}

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)
		metricsCount[m.Name()] += 1
	}

	require.Equal(t, map[string]int{
		"mysql.buffer_pool_operations": 7,
		"mysql.buffer_pool_pages":      6,
		"mysql.buffer_pool_size":       3,
		"mysql.commands":               6,
		"mysql.double_writes":          2,
		"mysql.handlers":               18,
		"mysql.locks":                  2,
		"mysql.log_operations":         3,
		"mysql.operations":             3,
		"mysql.page_operations":        3,
		"mysql.row_locks":              2,
		"mysql.row_operations":         4,
		"mysql.sorts":                  4,
		"mysql.threads":                4,
	}, metricsCount)
}

func (suite *MysqlIntegrationSuite) TestStartStop() {
	t := suite.T()
	container := mysqlContainer(t)
	defer container.Terminate(context.Background())

	sc := newMySQLScraper(zap.NewNop(), &Config{
		User:     "otel",
		Password: "otel",
		Endpoint: "127.0.0.1:3306",
	})

	// require scraper to connection to be open
	err := sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	require.False(t, sc.client.Closed())

	// require scraper to connection to be closed
	err = sc.shutdown(context.Background())
	require.NoError(t, err)
	require.True(t, sc.client.Closed())

	// require scraper to produce no error when closed again
	err = sc.shutdown(context.Background())
	require.NoError(t, err)
	require.True(t, sc.client.Closed())
}
