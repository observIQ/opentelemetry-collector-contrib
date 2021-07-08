package mysqlreceiver

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type client interface {
	getGlobalStats() ([]Stat, error)
	getInnodbStats() ([]Stat, error)
	Closed() bool
	Close() error
}

type mySQLClient struct {
	client *sql.DB
	closed bool
}

var _ client = (*mySQLClient)(nil)

type mySQLConfig struct {
	user string
	pass string
	addr string
	port int
}

func newMySQLClient(conf mySQLConfig) (client, error) {
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/", conf.user, conf.pass, conf.addr, conf.port)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}

	return &mySQLClient{
		client: db,
	}, nil
}

func (c *mySQLClient) getGlobalStats() ([]Stat, error) {
	query := "SHOW GLOBAL STATUS"
	result, err := execQuery(*c, query)
	if err != nil {
		return []Stat{}, err
	}
	return result, err
}

func (c *mySQLClient) getInnodbStats() ([]Stat, error) {
	query := "SELECT name, count FROM information_schema.innodb_metrics WHERE name LIKE '%buffer_pool_size%';"
	result, err := execQuery(*c, query)
	if err != nil {
		return []Stat{}, err
	}
	return result, err
}

type Stat struct {
	key   string
	value string
}

func execQuery(c mySQLClient, query string) ([]Stat, error) {
	rows, err := c.client.Query(query)
	if err != nil {
		log.Fatal(err)
		return []Stat{}, err
	}

	cols, err := rows.Columns()
	if err != nil {
		return []Stat{}, err
	}

	if len(cols) != 2 { // expected columns are 2: key and value.
		return []Stat{}, fmt.Errorf("expected 2 columns from query, got %d", len(cols))
	}

	rawResult := make([][]byte, len(cols))
	stats := []Stat{}

	dest := make([]interface{}, len(cols)) // A temporary interface{} slice
	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			return []Stat{}, err
		}

		stat := Stat{}
		for i, raw := range rawResult {
			if i == 0 {
				stat.key = string(raw)
			}
			if i == 1 {
				stat.value = string(raw)
			}
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (m *mySQLClient) Closed() bool {
	return m.closed
}

func (m *mySQLClient) Close() error {
	err := m.client.Close()
	if err != nil {
		return err
	}
	m.closed = true
	return nil
}
