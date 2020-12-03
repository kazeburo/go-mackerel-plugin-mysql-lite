package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jessevdk/go-flags"
	"github.com/kazeburo/go-mysqlflags"
)

// Version by Makefile
var version string

type opts struct {
	mysqlflags.MyOpts
	Timeout time.Duration `long:"timeout" default:"10s" description:"Timeout to connect mysql"`
	Version bool          `short:"v" long:"version" description:"Show version"`
}

type slave struct {
	IORunning  bool  `mysqlvar:"Slave_IO_Running"`
	SQLRunning bool  `mysqlvar:"Slave_SQL_Running"`
	Behind     int64 `mysqlvar:"Seconds_Behind_Master"`
}

type slaveMetric struct {
	IORunning  int64
	SQLRunning int64
	Behind     int64
}

type threads struct {
	Running   int64 `mysqlvar:"Threads_running"`
	Connected int64 `mysqlvar:"Threads_connected"`
	Cached    int64 `mysqlvar:"Threads_cached"`
}

type connections struct {
	Max       int64 `mysqlvar:"max_connections"`
	CacheSize int64 `mysqlvar:"thread_cache_size"`
}

func fetchSlaveStatus(db *sql.DB) (*slave, error) {
	var slaves []slave
	err := mysqlflags.Query(db, "SHOW SLAVE STATUS").Scan(&slaves)
	if err != nil {
		return nil, err
	}
	if len(slaves) == 0 {
		return &slave{}, nil
	}
	slave := slaves[0]
	return &slave, nil
}

func main() {
	os.Exit(_main())
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func _main() int {
	opts := opts{}
	psr := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	_, err := psr.Parse()
	if opts.Version {
		fmt.Fprintf(os.Stderr, "Version: %s\nCompiler: %s %s\n",
			version,
			runtime.Compiler,
			runtime.Version())
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	db, err := mysqlflags.OpenDB(opts.MyOpts, opts.Timeout, false)
	if err != nil {
		log.Printf("couldn't connect DB: %v", err)
		return 1
	}
	defer db.Close()

	var threads threads
	err = mysqlflags.Query(db, "SHOW GLOBAL STATUS").Scan(&threads)
	if err != nil {
		log.Printf("couldn't fetch status: %v", err)
		return 1
	}

	var connections connections
	err = mysqlflags.Query(db, "SHOW VARIABLES").Scan(&connections)
	if err != nil {
		log.Printf("couldn't fetch variables: %v", err)
		return 1
	}

	slave, err := fetchSlaveStatus(db)
	if err != nil {
		log.Printf("couldn't fetch slave status: %v", err)
		return 1
	}

	now := int32(time.Now().Unix())

	// slave status
	fmt.Printf("mysql-lite.replication-behind-master.second\t%d\t%d\n", slave.Behind, now)
	fmt.Printf("mysql-lite.replication-threads.io\t%d\t%d\n", btoi(slave.IORunning), now)
	fmt.Printf("mysql-lite.replication-threads.sql\t%d\t%d\n", btoi(slave.SQLRunning), now)

	// thread
	fmt.Printf("mysql-lite.threads.running\t%d\t%d\n", threads.Running, now)
	fmt.Printf("mysql-lite.threads.connected\t%d\t%d\n", threads.Connected, now)
	fmt.Printf("mysql-lite.threads.cached\t%d\t%d\n", threads.Cached, now)
	fmt.Printf("mysql-lite.threads.max-connections\t%d\t%d\n", connections.Max, now)
	fmt.Printf("mysql-lite.threads.cache-size\t%d\t%d\n", connections.CacheSize, now)

	fmt.Printf("mysql-lite.connections.utilization\t%f\t%d\n", float64(threads.Connected*100)/float64(connections.Max), now)

	return 0
}
