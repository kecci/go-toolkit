package sql

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gitlab.com/cms-tech/be-toolkit/lib/log"
)

type DB struct {
	// Master defines db operation that will be performed against master DB
	Master

	// Follower defines db operation that will be performed against follower DB
	Follower
	master   *sqlx.DB
	follower *sqlx.DB

	// driver define the base driver used. like postgres or mysql. nrpostgres will be converted as postgres
	driver string

	defaultTimeout time.Duration
}

type DBConfig struct {
	Driver                string        `json:"driver" yaml:"driver"`
	MasterDSN             string        `json:"master" yaml:"master"`
	FollowerDSN           string        `json:"follower" yaml:"follower"`
	MaxOpenConnections    int           `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConnections    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnectionMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`

	// number of retry during Connect
	// won't be used if `NoPingOnOpen`=true
	Retry int `json:"retry" yaml:"retry"`

	// no Ping when openning DB connection, useful if we don't care whether the server is up or not
	NoPingOnOpen bool `json:"no_ping_on_open" yaml:"no_ping_on_open"`
}

// Master defines operation that will be executed to master DB
type Master interface {
	Exec(query string, args ...interface{}) (sql.Result, error)

	// ExecContext use master database to exec query
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Begin transaction on master DB
	Begin() (*sql.Tx, error)

	// BeginTx begins transaction on master DB
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	// Rebind a query from the default bindtype (QUESTION) to the target bindtype.
	Rebind(sql string) string

	// NamedExec do named exec on master DB
	NamedExec(query string, arg interface{}) (sql.Result, error)

	// NamedExecContext do named exec on master DB
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)

	// BindNamed do BindNamed on master DB
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
}

// Follower defines operation that will be executed to follower DB
type Follower interface {
	// Get from follower database
	Get(dest interface{}, query string, args ...interface{}) error

	// Select from follower database
	Select(dest interface{}, query string, args ...interface{}) error

	// Query from follower database
	Query(query string, args ...interface{}) (*sql.Rows, error)

	// QueryRow executes QueryRow against follower DB
	QueryRow(query string, args ...interface{}) *sql.Row

	// NamedQuery do named query on follower DB
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)

	// GetContext from sql database
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// SelectContext from sql database
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	// QueryContext from sql database
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// QueryRowContext from sql database
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	// QueryxContext queries the database and returns an *sqlx.Rows. Any placeholder parameters are replaced with supplied args.
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)

	// QueryRowxContext queries the database and returns an *sqlx.Row. Any placeholder parameters are replaced with supplied args.
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row

	// NamedQueryContext do named query on follower DB
	NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
}

// WriteStatement is statement interface mean to be executed on Master DB.
// it only contains write operation
type WriteStatement interface {
	// ExecContext executes a prepared statement with the given arguments and returns a Result summarizing the effect of the statement.
	ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)

	// Close closes the statement.
	Close() error
}

// ReadStatement is statement interface  mean to be executed on Follower DB.
// It only contains read operation
type ReadStatement interface {
	// GetContext using the prepared statement.
	// Any placeholder parameters are replaced with supplied args.
	// An error is returned if the result set is empty.
	GetContext(ctx context.Context, dest interface{}, args ...interface{}) error

	// SelectContext using the prepared statement.
	// Any placeholder parameters are replaced with supplied args.
	SelectContext(ctx context.Context, dest interface{}, args ...interface{}) error

	// QueryContext from sql database
	QueryContext(ctx context.Context, arg ...interface{}) (*sql.Rows, error)

	// QueryRowContext from sql database
	QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row

	// QueryRowxContext queries the database and returns an *sqlx.Row.
	// Any placeholder parameters are replaced with supplied args.
	QueryRowxContext(ctx context.Context, args ...interface{}) *sqlx.Row

	// QueryxContext queries the database and returns an *sqlx.Rows.
	// Any placeholder parameters are replaced with supplied args.
	QueryxContext(ctx context.Context, args ...interface{}) (*sqlx.Rows, error)

	// Close closes the statement.
	Close() error
}

// NewFromDB creates *sqldb.DB from the existing *sql.DB.
//
// It can be used if we already have the *sql.DB object, usually during the test
func NewFromDB(masterDB *sql.DB, followerDB *sql.DB, driverName string) *DB {
	db := newFromSqlxDB(sqlx.NewDb(masterDB, driverName), sqlx.NewDb(followerDB, driverName))
	db.insertDriver(driverName)
	return db
}

func newFromSqlxDB(masterDB, followerDB *sqlx.DB) *DB {
	return &DB{
		Master:         masterDB,
		Follower:       followerDB,
		master:         masterDB,
		follower:       followerDB,
		defaultTimeout: 3 * time.Second,
	}
}

// Connect to kothak sql database object
func Connect(ctx context.Context, cfg DBConfig) (*DB, error) {
	masterdb, err := openOrConnect(ctx, cfg.Driver, cfg.MasterDSN, cfg.Retry, cfg.NoPingOnOpen)
	if err != nil {
		return nil, err
	}

	var followerdb *sqlx.DB

	if cfg.FollowerDSN != "" {
		followerdb, err = openOrConnect(ctx, cfg.Driver, cfg.FollowerDSN, cfg.Retry, cfg.NoPingOnOpen)
		if err != nil {
			return nil, err
		}
	} else { // if followerDSN is not configured, we use master DB as follower DB
		followerdb = masterdb
	}

	db := newFromSqlxDB(masterdb, followerdb)
	db.insertDriver(cfg.Driver)

	if cfg.MaxIdleConnections > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConnections)
	}
	if cfg.MaxOpenConnections > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConnections)
	}

	if cfg.ConnectionMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnectionMaxLifetime)
	}
	return db, nil
}

// PrepareWrite creates a prepared statement for write queries.
// The statement will be executed on Master DB
func (db *DB) PrepareWrite(ctx context.Context, query string) (WriteStatement, error) {
	return db.master.PreparexContext(ctx, query)
}

// PrepareRead creates a prepared statement for read queries.
// The statement will be executed on Follower DB
func (db *DB) PrepareRead(ctx context.Context, query string) (ReadStatement, error) {
	return db.follower.PreparexContext(ctx, query)
}

// Ping to sql database
func (db *DB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), db.defaultTimeout)
	defer cancel()
	return db.PingContext(ctx)
}

// PingContext function
func (db *DB) PingContext(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		errCh <- db.master.PingContext(ctx)
	}()

	go func() {
		errCh <- db.follower.PingContext(ctx)
	}()

	for i := 0; i < 2; i++ {
		err := <-errCh
		if err != nil {
			return err
		}
	}

	return nil
}

// GetMaster get master DB of sqldb
func (db *DB) GetMaster() *sqlx.DB {
	return db.master
}

// GetFollower return follower db
func (db *DB) GetFollower() *sqlx.DB {
	return db.follower
}

// SetMaxIdleConns to sql database
func (db *DB) SetMaxIdleConns(n int) {
	db.master.SetMaxIdleConns(n)
	db.follower.SetMaxIdleConns(n)
}

// SetMaxOpenConns to sql database
func (db *DB) SetMaxOpenConns(n int) {
	db.master.SetMaxOpenConns(n)
	db.follower.SetMaxOpenConns(n)
}

// SetConnMaxLifetime to sql database
func (db *DB) SetConnMaxLifetime(t time.Duration) {
	db.master.SetConnMaxLifetime(t)
	db.follower.SetConnMaxLifetime(t)
}

// insertDriver will set db module driver with base driver by check type of database.
// Currently only check for postgres and mysql
func (db *DB) insertDriver(driver string) {
	if driver == "nrpostgres" {
		db.driver = "postgres"
	} else if driver == "nrmysql" {
		db.driver = "mysql"
	} else {
		db.driver = driver
	}
}

// Rebind will do usual Rebind by driverName param in db.
// Please use this rather than Rebind in GetMaster() or GetFollower() to make sure the rebind is correct, especially if you use newrelic
func (db *DB) Rebind(query string) string {
	return sqlx.Rebind(sqlx.BindType(db.driver), query)
}

// BindNamed will do usual BindNamed by driverName param in db.
// Please use this rather than BindNamed in GetMaster() or GetFollower() to make sure the rebind is correct, especially if you use newrelic
func (db *DB) BindNamed(query string, arg interface{}) (string, []interface{}, error) {
	return sqlx.BindNamed(sqlx.BindType(db.driver), query, arg)
}

// openOrConnect will do one these things based on the value of `noPing` argument
// - true  : call sqlx.Open which only creating sqlx.DB object
// - false : call sqlx.Connect which is sqlx.Open + Ping to DB.
//		     if the Ping failed, we retry it for the configured `retry` argument.
func openOrConnect(ctx context.Context, driver, dsn string, retry int, noPing bool) (*sqlx.DB, error) {
	if noPing {
		return sqlx.Open(driver, dsn)
	}
	return connectWithRetry(ctx, driver, dsn, retry)
}

func connectWithRetry(ctx context.Context, driver, dsn string, retry int) (*sqlx.DB, error) {
	var (
		db        *sqlx.DB
		err       error
		noPassDSN = getNoPassDSN(dsn)
	)

	if retry <= 0 {
		retry = 1
	}

	for x := 0; x < retry; x++ {
		db, err = connect(ctx, driver, dsn)
		if err == nil {
			return db, nil
		}
		log.Warnf("SQLDB: failed to connect to %s with error %s", noPassDSN, err.Error())

		if x+1 < retry {
			// continue with condition
			log.Warnf("sqldb: retrying to connect to %s. Retry: %d", noPassDSN, x+1)
			// sleep for 3 secs in every retries
			time.Sleep(time.Second * 3)
		}
	}

	log.Errorf("sqldb: retry time exhausted, cannot connect to database: %s", err.Error())
	err = fmt.Errorf("failed connect to database: %s", err.Error())
	return nil, err
}

func connect(ctx context.Context, driver, dsn string) (*sqlx.DB, error) {
	return sqlx.ConnectContext(ctx, driver, dsn)
}

var dsnPasswordPattern = regexp.MustCompile(`(password=[^\s]*\s*|$)|(:[^/][^@]*)`)

func getNoPassDSN(dsn string) string {
	return strings.TrimSpace(dsnPasswordPattern.ReplaceAllString(dsn, ""))
}
