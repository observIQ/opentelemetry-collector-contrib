// Copyright 2021, observIQ
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
