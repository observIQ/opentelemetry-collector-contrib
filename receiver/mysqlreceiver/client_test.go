package mysqlreceiver

import (
	"bufio"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var _ client = (*fakeClient)(nil)

type fakeClient struct {
}

func newFakeClient() *fakeClient {
	return &fakeClient{}
}

func readFile(fname string) ([]*Stat, error) {
	var stats = []*Stat{}
	file, err := os.Open(path.Join("testdata", fname+".txt"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.Split(scanner.Text(), "\t")
		stats = append(stats, &Stat{key: text[0], value: text[1]})

	}
	return stats, nil
}

func (c *fakeClient) getGlobalStats() ([]*Stat, error) {
	return readFile("global_stats")
}

func TestGlobalStats(t *testing.T) {
	client := newFakeClient()
	res, err := client.getGlobalStats()
	require.Nil(t, err)
	require.NotNil(t, res)
	require.False(t, client.Closed())
}

func (c *fakeClient) getInnodbStats() ([]*Stat, error) {
	return readFile("innodb_stats")
}

func TestInnodbStats(t *testing.T) {
	client := newFakeClient()
	res, err := client.getInnodbStats()
	require.Nil(t, err)
	require.NotNil(t, res)
	require.False(t, client.Closed())

}

func (c *fakeClient) Closed() bool {
	return false
}

func (c *fakeClient) Close() error {
	return nil
}
