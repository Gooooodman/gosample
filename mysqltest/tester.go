package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

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

// 向历史数据表中批量插入数据
func batchInsertDataHistory(db *sql.DB, tablename string, count int32, batch int32) (int32, error) {
	sqlpref := `INSERT INTO %s VALUES `
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(sqlpref, tablename))
	var i, ibatch int32
	sformat := "2006-01-02 15:04:05.000 "
	stime := time.Now().Format(sformat)
	vformat := "('%s','%s', 1, '%s', 1234, 1, 2)%s "
	for i = 0; i < count; i++ {
		nodeid := fmt.Sprintf("%s:%d", tablename, i)

		ibatch++
		if ibatch == count {
			buffer.WriteString(fmt.Sprintf(vformat, nodeid, stime, stime, " "))
		} else if ibatch%batch != 0 {
			buffer.WriteString(fmt.Sprintf(vformat, nodeid, stime, stime, ","))
		} else {
			ibatch = 0
			buffer.WriteString(fmt.Sprintf(vformat, nodeid, stime, stime, " "))
			//log.Println("<<<:", buffer.String())

			_, err := db.Exec(buffer.String())
			if err != nil {
				return i, err
			}
			buffer.Reset()
			buffer.WriteString(fmt.Sprintf(sqlpref, tablename))
		}
	}
	if ibatch > 0 {
		//fmt.Println(">>>:", buffer.String())
		_, err := db.Exec(buffer.String())
		if err != nil {
			return i, err
		}
	}
	return i, nil
}

// 插入任务，每6秒钟插入5000条数据
func insertTask(db *sql.DB, taskno int, done <-chan bool) error {
	tablename := fmt.Sprintf("t_data_unit%d", taskno)
	// 初始化数据库表结构
	err := initDbSchema(db, tablename)
	if err != nil {
		return errors.New(fmt.Sprintf("!任务[%02d]:创建数据库表[%s]出错。\n\t\t错误栈信息[%s]", taskno, tablename, err.Error()))
	}
	timer := time.NewTicker(6 * time.Second)
	for {
		select {
		case <-done:
			log.Printf("-任务[%00d]:开始退出。", taskno)
			return nil
		case current := <-timer.C:
			begin := time.Now()
			_, err := batchInsertDataHistory(db, tablename, 5000, 500)
			if err != nil {
				return errors.New(fmt.Sprintf("!任务[%02d]:执行批量插入数据时出错。\n\t\t错误栈信息[%s]", taskno, err.Error()))
			}
			sformat := "2006-01-02 15:04:05.000 "
			diff := time.Now().Sub(begin)
			log.Printf("*任务[%02d][%s]:成功插入5000条数据，耗时[%6.3f]秒", taskno, current.Format(sformat), diff.Seconds())
		}
	}
	return nil
}

// 多个插入任务
func multiInserTask(count int, start int) {
	dbconf := &DbConfig{DriverName: "mysql", Username: "scada", Password: "scada", Host: "172.26.17.53", Port: 3306, DBName: "scada"}
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
	log.Println("-请输入命令[start, stop]")
	for running {
		data, _, _ := reader.ReadLine()
		command := string(data)
		switch command {
		case "stop":
			log.Println("-执行Stop方法开始。")
			for _, done := range dones {
				done <- true
			}
			time.Sleep(2 * time.Second)
			running = false
			log.Println("-执行Stop方法结束。")
		case "start":
			for i := start; i < count+1; i++ {
				go func(taskno int) {
					err := insertTask(db, taskno, dones[taskno-1])
					if err != nil {
						log.Println(err.Error())
					}
				}(i)
			}
		default:
			log.Printf("-命令%s不存在，请收入[start, stop]。", command)
		}
	}
}

// 主函数
func main() {
	taskTypePtr := flag.String("type", "insert", "任务类型[insert, query]")
	startPtr := flag.Int("start", 1, "任务起始编号（整型）")
	countPtr := flag.Int("count", 1, "任务个数（整型）")
	flag.Parse()

	taskType := *taskTypePtr
	start := *startPt
	count := *countPtr
	if taskType != "insert" && taskType != "query" {
		log.Println("参数[type]值不正确，该参数取值为[insert, query]")
		return
	}

	if taskType == "insert" {
		multiInserTask(count, start)
	}
}
