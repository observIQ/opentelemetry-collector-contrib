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
	"database/sql"
	"fmt"

	// registers the mysql driver
	_ "github.com/go-sql-driver/mysql"
)

type client interface {
	getGlobalStats() ([]*Stat, error)
	getInnodbStats() ([]*Stat, error)
	Closed() bool
	Close() error
}

type mySQLClient struct {
	client *sql.DB
	closed bool
}

var _ client = (*mySQLClient)(nil)

type mySQLConfig struct {
	user     string
	pass     string
	endpoint string
}

func newMySQLClient(conf mySQLConfig) (*mySQLClient, error) {
	connStr := fmt.Sprintf("%s:%s@tcp(%s)/", conf.user, conf.pass, conf.endpoint)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}

	return &mySQLClient{
		client: db,
	}, nil
}

// getGlobalStats queries the db for global status metrics.
func (c *mySQLClient) getGlobalStats() ([]*Stat, error) {
	query := "SHOW GLOBAL STATUS;"
	result, err := Query(*c, query)
	if err != nil {
		return nil, err
	}
	return result, err
}

// getInnodbStats queries the db for innodb metrics.
func (c *mySQLClient) getInnodbStats() ([]*Stat, error) {
	query := "SELECT name, count FROM information_schema.innodb_metrics WHERE name LIKE '%buffer_pool_size%';"
	result, err := Query(*c, query)
	if err != nil {
		return nil, err
	}
	return result, err
}

type Stat struct {
	key   string
	value string
}

func Query(c mySQLClient, query string) ([]*Stat, error) {
	rows, err := c.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stats := make([]*Stat, 0)
	for rows.Next() {
		var stat Stat
		if err := rows.Scan(&stat.key, &stat.value); err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

func (c *mySQLClient) Closed() bool {
	return c.closed
}

func (c *mySQLClient) Close() error {
	err := c.client.Close()
	if err != nil {
		return err
	}
	c.closed = true
	return nil
}
