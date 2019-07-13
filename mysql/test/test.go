package main

import (
	_ "database/sql"
	"fmt"
	"github.com/nothollyhigh/kiss/mysql"
	"sync"
)

const testMysqlDBConn = "root:123qwe@tcp(localhost:3306)/tmpdb"

var (
	testMysqlInited = false

	dbMysql *mysql.MysqlMgr
)

/*
// init sql
create database tmpdb
drop table tmptab
create table tmptab (
	id bigint primary key auto_increment,
    name  varchar(32) not null,
    remark varchar(64) not null
);
alter table tmptab auto_increment=100000000;
*/
func testInitMysql() {
	if !testMysqlInited {
		testMysqlInited = true

		dbMysql = mysql.NewMysqlMgr(map[string][]mysql.MysqlConf{
			"master": []mysql.MysqlConf{
				mysql.MysqlConf{
					ConnString:        testMysqlDBConn,
					PoolSize:          10,
					KeepaliveInterval: 2, //300
				},
			},
		})
	}
}

func main() {
	testInitMysql()
	wg := sync.WaitGroup{}
	for i := 0; i < 1; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				func() {
					if db := dbMysql.Get("master").DB(); db != nil {
						name := fmt.Sprintf("name_%d_%d", idx, j)
						remark := fmt.Sprintf("remark_%d_%d", idx, j)
						ret, err := db.Exec(`insert into tmptab (name, remark) values(?, ?)`, name, remark)
						if err != nil {
							fmt.Printf("Insert: (%s, %s), error: %v\n", name, remark, err)
							return
						}

						id, err := ret.LastInsertId()
						ret, err = db.Exec(`update tmptab set remark=? where id=?`, remark+"_update", id)
						if err != nil {
							fmt.Printf("update: (%s, %s, %v), error: %v\n", name, remark, id, err)
							return
						}
						eff, err := ret.RowsAffected()
						if err != nil {
							fmt.Printf("Insert: (%s, %s), error: %v\n", name, remark, err)
							return
						}

						retname, retremark := "", ""
						if err := db.QueryRow(`select name,remark from tmpdb.tmptab where name=?`, name).Scan(&retname, &retremark); err != nil || !(name == retname && remark+"_update" == retremark) {
							fmt.Printf("Find: (%s, %s), error: %v, equal: %v\n", retname, retremark, err, name == retname && remark == retremark)
						}

						fmt.Printf("insert success: %v, %v, %v, %v\n", id, eff, name, remark+"_update")
					}
				}()
			}
		}()
	}

	wg.Wait()
}
