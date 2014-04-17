package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const table_prefix = "t_data_unit"

var debug bool = false

// 数据库配置参数类
type DbConfig struct {
	DriverName string
	Username   string
	Password   string
	Host       string
	Port       int
	DBName     string
}

// 构造sql/DB的Open方法需要的DataSourceName字符串
func (dbconf *DbConfig) DataSourceName() string {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbconf.Username, dbconf.Password, dbconf.Host, dbconf.Port, dbconf.DBName)
	//fmt.Println(dataSourceName)
	return dataSourceName
}

// 判断表是否存在
func existTable(db *sql.DB, dbname string, tablename string) (bool, error) {
	sql := `SELECT COUNT(*) 
            FROM information_schema.tables 
            WHERE table_schema=? AND table_name=?`
	stmt, err := db.Prepare(sql)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var count int
	err = stmt.QueryRow(dbname, tablename).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

// 初始化数据库结构
func initDbSchema(db *sql.DB, tablename string) error {
	dbname := "scada"
	exist, err := existTable(db, dbname, tablename)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	sql := `CREATE TABLE %s(
                node_id             VARBINARY(128), 
                value_time          BINARY(24),
                sub_index           SMALLINT UNSIGNED,
                value_time2         BINARY(24),
                node_value          VARBINARY(64),
                value_quality       SMALLINT UNSIGNED,
                value_type          SMALLINT UNSIGNED,
                PRIMARY KEY(node_id, sub_index, value_time)
            );`
	_, err = db.Exec(fmt.Sprintf(sql, tablename))
	return err
}

// 从历史库查询数据
func queryDataHistory(db *sql.DB, taskno int) (int32, error) {
	sql := fmt.Sprintf("SELECT * FROM %s%d WHERE node_id=? AND value_time >= ? AND value_time < ? LIMIT 100", table_prefix, taskno)
	nodeid := fmt.Sprintf("%s%d:%d", table_prefix, taskno, rand.Intn(5000))
	begin := time.Now().Add(-3600 * time.Second)
	end := begin.Add(5 * time.Minute)
	if rand.Intn(2) == 1 {
		end = begin.Add(24 * time.Hour)
	}
	rows, err := db.Query(sql, nodeid, begin, end)
	if err != nil {
		return 0, err
	}
	var count int32 = 0
	for rows.Next() {
		count++
	}
	return count, nil
}

// 向历史数据表中批量插入数据
func batchInsertDataHistory(db *sql.DB, taskno int, count int, batch int) (int, error) {
	sqlpref := `INSERT INTO %s%d VALUES `
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(sqlpref, table_prefix, taskno))
	var i, ibatch int
	sformat := "2006-01-02 15:04:05.000 "
	stime := time.Now().Format(sformat)
	vformat := "('%s','%s', 1, '%s', 1234, 1, 2)%s "
	for i = 0; i < count; i++ {
		nodeid := fmt.Sprintf("%s%d:%d", table_prefix, taskno, i)

		ibatch++
		if i == count-1 {
			buffer.WriteString(fmt.Sprintf(vformat, nodeid, stime, stime, " "))
		} else if ibatch%batch != 0 {
			buffer.WriteString(fmt.Sprintf(vformat, nodeid, stime, stime, ","))
		} else {
			ibatch = 0
			buffer.WriteString(fmt.Sprintf(vformat, nodeid, stime, stime, " "))
			if debug {
				fmt.Println("-插入任务SQL(0):", buffer.String())
			}
			_, err := db.Exec(buffer.String())
			if err != nil {
				return i, err
			}
			buffer.Reset()
			buffer.WriteString(fmt.Sprintf(sqlpref, table_prefix, taskno))
		}
	}
	if ibatch > 0 {
		if debug {
			fmt.Println("-插入任务SQL(1):", buffer.String())
		}
		_, err := db.Exec(buffer.String())
		if err != nil {
			return i, err
		}
	}
	return i, nil
}

// 查询任务，每一分钟查询一次
func queryTask(db *sql.DB, taskno int, done <-chan bool) error {
	timer := time.NewTicker(time.Minute)
	for {
		select {
		case <-done:
			if debug {
				log.Printf("-查询任务[%02d]:开始退出。", taskno)
			}
			return nil
		case current := <-timer.C:
			begin := time.Now()
			count, err := queryDataHistory(db, taskno)
			if err != nil {
				return errors.New(fmt.Sprintf("!查询任务[%02d]:执行查询数据时出错。\n\t\t错误栈信息[%s]", taskno, err.Error()))
			}
			sformat := "2006-01-02 15:04:05.000 "
			diff := time.Now().Sub(begin)
			log.Printf("*查询任务[%02d][%s]:成功查询[%04d]条数据，耗时[%6.3f]秒", taskno, current.Format(sformat), count, diff.Seconds())
		}
	}
	return nil
}

// 插入任务，每6秒钟插入5000条数据
func insertTask(db *sql.DB, taskno int, done <-chan bool, batch int) error {
	tablename := fmt.Sprintf("t_data_unit%d", taskno)
	// 初始化数据库表结构
	err := initDbSchema(db, tablename)
	if err != nil {
		return errors.New(fmt.Sprintf("!插入任务[%02d]:创建数据库表[%s]出错。\n\t\t错误栈信息[%s]", taskno, tablename, err.Error()))
	}
	timer := time.NewTicker(6 * time.Second)
	for {
		select {
		case <-done:
			if debug {
				log.Printf("-插入任务[%00d]:开始退出。", taskno)
			}
			return nil
		case current := <-timer.C:
			begin := time.Now()
			_, err := batchInsertDataHistory(db, taskno, 5000, batch)
			if err != nil {
				return errors.New(fmt.Sprintf("!插入任务[%02d]:执行批量插入数据时出错。\n\t\t错误栈信息[%s]", taskno, err.Error()))
			}
			sformat := "2006-01-02 15:04:05.000 "
			diff := time.Now().Sub(begin)
			log.Printf("*插入任务[%02d][%s]:成功插入5000条数据，耗时[%6.3f]秒", taskno, current.Format(sformat), diff.Seconds())
		}
	}
	return nil
}

// 多个任务
func multiTask(count int, start int, tasktype string, host string, batch int) {
	dbconf := &DbConfig{DriverName: "mysql", Username: "scada", Password: "scada", Host: host, Port: 3306, DBName: "scada"}
	db, err := sql.Open(dbconf.DriverName, dbconf.DataSourceName())
	if err != nil {
		log.Fatalln(err)
		panic(err.Error())
	}
	defer db.Close()

	dones := make([]chan bool, count)
	for i, _ := range dones {
		dones[i] = make(chan bool, 1)
	}
	running := true
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("-请输入命令[start, stop]")
	for running {
		data, _, _ := reader.ReadLine()
		command := string(data)
		switch command {
		case "stop":
			if debug {
				fmt.Println("-执行Stop方法开始。")
			}
			for _, done := range dones {
				done <- true
			}
			time.Sleep(2 * time.Second)
			running = false
			if debug {
				fmt.Println("-执行Stop方法结束。")
			}
		case "start":
			for i := start; i < count+1; i++ {
				go func(taskno int) {
					var err error = nil
					if "insert" == tasktype {
						err = insertTask(db, taskno, dones[taskno-1], batch)
					} else {
						err = queryTask(db, taskno, dones[taskno-1])
					}
					if err != nil {
						log.Println(err.Error())
					}
				}(i)
			}
		default:
			fmt.Printf("-命令%s不存在，请收入[start, stop]。", command)
		}
	}
}

// 主函数
func main() {
	tasktypePtr := flag.String("type", "insert", "任务类型[insert, query]")
	startPtr := flag.Int("start", 1, "任务起始编号（整型）")
	countPtr := flag.Int("count", 1, "任务个数（整型）")
	debugPtr := flag.Bool("debug", false, "Debug模式")
	logfilePtr := flag.String("logfile", "/home/scada/mysqltest.log", "log文件位置")
	hostPtr := flag.String("host", "127.0.0.1", "host地址")
	batchPtr := flag.Int("batch", 500, "每批插入记录数")
	flag.Parse()

	tasktype := *tasktypePtr
	start := *startPtr
	count := *countPtr
	debug = *debugPtr
	host := *hostPtr
	batch := *batchPtr
	if tasktype != "insert" && tasktype != "query" {
		fmt.Println("参数[type]值不正确，该参数取值为[insert, query]")
		return
	}
	logfile, err := os.OpenFile(*logfilePtr, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		fmt.Println("打开日志文件[%s]出错", *logfilePtr)
		return
	}
	defer logfile.Close()
	log.SetOutput(logfile)

	multiTask(count, start, tasktype, host, batch)
}
