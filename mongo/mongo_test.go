package mongo

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"testing"
	"time"
)

//const testMongoDBConn = "mongodb://127.0.0.1:27017?maxPoolSize=10/admin"
const testMongoDBConn = "mongodb://127.0.0.1:27017/admin"

var (
	testMongoInited = false

	dbMongo *Mongo
)

func testInitMongo() {
	if !testMongoInited {
		testMongoInited = true

		dbMongo = NewMongo(MongoConf{
			//ConnString:        testMongoDBConn,
			Addrs:             []string{"127.0.0.1:27017"},
			Database:          "admin",
			PoolSize:          100,
			DialTimeout:       5,
			SocketTimeout:     10,
			KeepaliveInterval: 2, //300

			// http://www.mongoing.com/archives/1723
			Safe: &mgo.Safe{
				J:     true,       //true:写入落到磁盘才会返回|false:不等待落到磁盘|此项保证落到磁盘
				W:     1,          //0:不会getLastError|1:主节点成功写入到内存|此项保证正确写入
				WMode: "majority", //"majority":多节点写入|此项保证一致性|如果我们是单节点不需要这只此项
			},
		})
	}
}

func TestMongo(t *testing.T) {
	testInitMongo()

	wg := &sync.WaitGroup{}

	type Doc struct {
		Name  string
		Age   int
		Sex   int
		Phone string
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		doc := Doc{
			Name:  fmt.Sprintf("name_%d", i),
			Age:   10 + i%60,
			Sex:   i % 2,
			Phone: "15012345678",
		}
		go func() {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				t0 := time.Now()
				if true { //doc.Sex%2 == 0 {
					func() {
						sess, err := dbMongo.SessionWithTimeout(time.Second)
						if err != nil {
							fmt.Printf("time used: %v, insert %v error: %v\n", time.Since(t0), doc.Name, err)
							return
						}
						sessCopy := sess.Clone()
						//if j < 10 {
						defer sessCopy.Close()
						//}
						err = sessCopy.DB("dbtestinsert").C("test").Insert(doc)
						//if err != nil {
						fmt.Printf("time used: %v, insert %v error: %v\n", time.Since(t0), doc.Name, err)

						//}
					}()
				} else {
					err := dbMongo.Insert("dbtestinsert", "test", doc)
					//if err != nil {
					fmt.Printf("time used: %v, insert %v error: %v\n", time.Since(t0), doc.Name, err)
					//}
				}
			}
		}()
	}
	wg.Wait()
}
