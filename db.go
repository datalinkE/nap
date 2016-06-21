package sqlxentrypoint

import (
	"github.com/jmoiron/sqlx"
	"strings"
	"sync/atomic"
)

// DB is a entry point for logical database with multiple underlying physical databases
// forming a single master multiple slaves topology.
type DB struct {
	pdbs  []*sqlx.DB // Physical databases
	count uint64     // Monotonically incrementing counter on each query
}

// Open concurrently opens each underlying physical db.
// dataSourceNames must be a semi-comma separated list of DSNs with the first
// one being used as the master and the rest as slaves.
func Open(driverName, dataSourceNames string) (*DB, error) {
	conns := strings.Split(dataSourceNames, ";")
	db := &DB{pdbs: make([]*sqlx.DB, len(conns))}

	err := scatter(len(db.pdbs), func(i int) (err error) {
		db.pdbs[i], err = sqlx.Open(driverName, conns[i])
		return err
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}

// Close closes all physical databases concurrently, releasing any open resources.
func (db *DB) Close() error {
	return scatter(len(db.pdbs), func(i int) error {
		return db.pdbs[i].Close()
	})
}

// Ping verifies if a connection to each physical database is still alive,
// establishing a connection if necessary.
func (db *DB) Ping() error {
	return scatter(len(db.pdbs), func(i int) error {
		return db.pdbs[i].Ping()
	})
}

// SetMaxIdleConns sets the maximum number of connections in the idle
// connection pool for each underlying physical db.
// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns then the
// new MaxIdleConns will be reduced to match the MaxOpenConns limit
// If n <= 0, no idle connections are retained.
func (db *DB) SetMaxIdleConns(n int) {
	for i := range db.pdbs {
		db.pdbs[i].SetMaxIdleConns(n)
	}
}

// SetMaxOpenConns sets the maximum number of open connections
// to each physical database.
// If MaxIdleConns is greater than 0 and the new MaxOpenConns
// is less than MaxIdleConns, then MaxIdleConns will be reduced to match
// the new MaxOpenConns limit. If n <= 0, then there is no limit on the number
// of open connections. The default is 0 (unlimited).
func (db *DB) SetMaxOpenConns(n int) {
	for i := range db.pdbs {
		db.pdbs[i].SetMaxOpenConns(n)
	}
}

// Slave returns one of the physical databases which is a slave
func (db *DB) Slave() *sqlx.DB {
	return db.pdbs[db.slave(len(db.pdbs))]
}

// Master returns the master physical database
func (db *DB) Master() *sqlx.DB {
	return db.pdbs[0]
}

func (db *DB) slave(n int) int {
	if n <= 1 {
		return 0
	}
	return int(1 + (atomic.AddUint64(&db.count, 1) % uint64(n-1)))
}
