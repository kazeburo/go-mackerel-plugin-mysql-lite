package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jessevdk/go-flags"
	"github.com/kazeburo/go-mysqlflags"
)

// Version by Makefile
var version string

type Opts struct {
	mysqlflags.MyOpts
	Timeout time.Duration `long:"timeout" default:"10s" description:"Timeout to connect mysql"`
	Version bool          `short:"v" long:"version" description:"Show version"`
}

func fetchStatus(db *sql.DB) (map[string]string, error) {
	return mysqlflags.QueryMapCol(db, "SHOW GLOBAL STATUS")
}

func fetchVariables(db *sql.DB) (map[string]string, error) {
	return mysqlflags.QueryMapCol(db, "SHOW VARIABLES")
}

func fetchSlaveStatus(db *sql.DB) (map[string]string, error) {
	rows, err := mysqlflags.QueryMapRows(db, "SHOW SLAVE STATUS")
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]string{
			"Slave_IO_Running":      "0",
			"Slave_SQL_Running":     "0",
			"Seconds_Behind_Master": "0",
		}, nil
	}
	row := rows[0]
	result := map[string]string{}
	v, ok := row["Seconds_Behind_Master"]
	if !ok {
		return nil, fmt.Errorf("No Seconds_Behind_Master in result")
	}
	result["Seconds_Behind_Master"] = v

	keys := []string{"Slave_IO_Running", "Slave_SQL_Running"}
	for _, k := range keys {
		v, ok := row[k]
		if !ok {
			return nil, fmt.Errorf("No %s in result", k)
		}
		switch v {
		case "Yes":
			result[k] = "1"
		default:
			result[k] = "0"
		}
	}

	return result, nil
}

func main() {
	os.Exit(_main())
}

func _main() int {
	opts := Opts{}
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

	st, err := fetchStatus(db)
	if err != nil {
		log.Printf("couldn't fetch status: %v", err)
		return 1
	}

	val, err := fetchVariables(db)
	if err != nil {
		log.Printf("couldn't fetch variables: %v", err)
		return 1
	}

	slaveSt, err := fetchSlaveStatus(db)
	if err != nil {
		log.Printf("couldn't fetch slave status: %v", err)
		return 1
	}

	now := int32(time.Now().Unix())

	// slave status
	fmt.Printf("mysql-lite.replication-behind-master.second\t%s\t%d\n", slaveSt["Seconds_Behind_Master"], now)
	fmt.Printf("mysql-lite.replication-threads.io\t%s\t%d\n", slaveSt["Slave_IO_Running"], now)
	fmt.Printf("mysql-lite.replication-threads.sql\t%s\t%d\n", slaveSt["Slave_SQL_Running"], now)

	// thread
	fmt.Printf("mysql-lite.threads.running\t%s\t%d\n", st["Threads_running"], now)
	fmt.Printf("mysql-lite.threads.connected\t%s\t%d\n", st["Threads_connected"], now)
	fmt.Printf("mysql-lite.threads.cached\t%s\t%d\n", st["Threads_cached"], now)
	fmt.Printf("mysql-lite.threads.max-connections\t%s\t%d\n", val["max_connections"], now)
	fmt.Printf("mysql-lite.threads.cache-size\t%s\t%d\n", val["thread_cache_size"], now)

	// connection utilization
	maxConnections, err := strconv.ParseFloat(val["max_connections"], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed parsing max_connections: %s\n", err)
		return 1
	}
	threadConnected, err := strconv.ParseFloat(st["Threads_connected"], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed parsing thread_connected: %s\n", err)
		return 1
	}
	fmt.Printf("mysql-lite.connections.utilization\t%f\t%d\n", threadConnected/maxConnections*100, now)

	return 0
}
