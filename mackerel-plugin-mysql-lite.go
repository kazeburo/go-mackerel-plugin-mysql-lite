package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"os"
	"time"
	"strconv"
)

type connectionOpts struct {
	Host string `short:"H" long:"host" default:"localhost" description:"Hostname"`
	Port string `short:"p" long:"port" default:"3306" description:"Port"`
	User string `short:"u" long:"user" default:"root" description:"Username"`
	Pass string `short:"P" long:"password" default:"" description:"Password"`
}


func fetchStatus(db mysql.Conn, stat map[string]string) error {
	rows, _, err := db.Query("SHOW GLOBAL STATUS")
	if err != nil {
		return err
	}

	for _, row := range rows {
		statusKey := row.Str(0)
		stat[statusKey] = row.Str(1)
	}
	return nil
}

func fetchVariables(db mysql.Conn, stat map[string]string) error {
	rows, _, err := db.Query("SHOW VARIABLES")
	if err != nil {
		return err
	}

	for _, row := range rows {
		statusKey := row.Str(0)
		stat[statusKey] = row.Str(1)
	}
	return nil
}

func fetchSlaveStatus(db mysql.Conn, stat map[string]string) error {
	rows, res, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return err
	}
	if len(rows) < 1 {
		stat["Slave_IO_Running"]  = "0"
		stat["Slave_SQL_Running"]  = "0"
		stat["Seconds_Behind_Master"] = "0"
		return nil
	}

	idxBehindMaster := res.Map("Seconds_Behind_Master")
	stat["Seconds_Behind_Master"] = rows[0].Str(idxBehindMaster)
	
	keys := []string{"Slave_IO_Running","Slave_SQL_Running"}
	for _, key := range keys {
		idx := res.Map(key)
		val := rows[0].Str(idx)
		if val == "Yes" {
			stat[key] = "1"
		} else {
			stat[key] = "0"
		}
	}
	return nil
}


func main() {
	os.Exit(_main())
}

func _main() (st int) {
	opts := connectionOpts{}
	psr := flags.NewParser(&opts, flags.Default)
	_, err := psr.Parse()
	if err != nil {
		os.Exit(1)
	}

	db := mysql.New("tcp", "", fmt.Sprintf("%s:%s", opts.Host, opts.Port), opts.User, opts.Pass, "")
	err = db.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't connect DB\n")
		return
	}
	defer db.Close()

	stat := make(map[string]string)
	err = fetchStatus(db, stat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't fetch status %s\n",err)
		return
	}

	variable := make(map[string]string)
	err = fetchVariables(db, variable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't fetch variables %s\n",err)
		return
	}

	slaveStat := make(map[string]string)
	err = fetchSlaveStatus(db, slaveStat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't fetch slave statuss %s\n",err)
		return
	}

	now := int32(time.Now().Unix())

	// slave status
	fmt.Printf("mysql-lite.replication-behind-master.second\t%s\t%d\n", slaveStat["Seconds_Behind_Master"], now)
	fmt.Printf("mysql-lite.replication-threads.io\t%s\t%d\n", slaveStat["Slave_IO_Running"], now)
	fmt.Printf("mysql-lite.replication-threads.sql\t%s\t%d\n", slaveStat["Slave_SQL_Running"], now)

	// thread
	fmt.Printf("mysql-lite.threads.running\t%s\t%d\n", stat["Threads_running"], now)
	fmt.Printf("mysql-lite.threads.connected\t%s\t%d\n", stat["Threads_connected"], now)
	fmt.Printf("mysql-lite.threads.cached\t%s\t%d\n", stat["Threads_cached"], now)
	fmt.Printf("mysql-lite.threads.max-connections\t%s\t%d\n", variable["max_connections"], now)
	fmt.Printf("mysql-lite.threads.cache-size\t%s\t%d\n", variable["thread_cache_size"], now)
	
	// connection utilization
	max_connections, err := strconv.ParseFloat(variable["max_connections"], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed parsing max_connections: %s\n",err)
		return
	}
	thread_connected, err := strconv.ParseFloat(stat["Threads_connected"], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed parsing thread_connected: %s\n",err)
		return
	}
	fmt.Printf("mysql-lite.connections.utilization\t%f\t%d\n",thread_connected/max_connections*100 , now)

	st = 0
	return
}



