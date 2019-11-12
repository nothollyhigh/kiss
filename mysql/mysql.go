package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"sort"
	"strings"
	"time"
)

type Config struct {
	ID                string `json:"ID"`
	ConnString        string `json:"ConnString"` //user:password@tcp(localhost:5555)/dbname
	EncryptKey        string `json:"EncryptKey"`
	PoolSize          int    `json:"PoolSize"`
	IdleSize          int    `json:"IdleSize"`
	KeepaliveInterval int    `json:"KeepaliveInterval"`
	keepaliveInterval time.Duration
}

type Mysql struct {
	db     *gorm.DB
	ticker *time.Ticker
	Conf   Config
}

func (msql *Mysql) OrmDB() *gorm.DB {
	return msql.db
}

func (msql *Mysql) DB() *sql.DB {
	return msql.db.DB()
}

func (msql *Mysql) Close() error {
	return msql.db.Close()
}

func New(conf Config) *Mysql {
	if conf.ConnString == "" {
		log.Fatal("msyql.New Failed: invalid ConnString")
	}

	if conf.KeepaliveInterval > 0 {
		conf.keepaliveInterval = time.Second * time.Duration(conf.KeepaliveInterval)
	} else {
		conf.keepaliveInterval = time.Second * 300
	}

	log.Info("msyql.New Connect To Mysql ...")

	// db, err := sql.Open("mysql", conf.ConnString)
	db, err := gorm.Open("mysql", conf.ConnString)
	if err != nil {
		log.Fatal("msyql.New sql.Open Failed: %v", err)
	}

	err = db.DB().Ping()
	if err != nil {
		log.Fatal("msyql.New Ping() Failed: %v", err)
	}

	if conf.PoolSize > 0 {
		db.DB().SetMaxOpenConns(conf.PoolSize)
	}
	if conf.IdleSize > 0 {
		db.DB().SetMaxIdleConns(conf.IdleSize)
	}

	msql := &Mysql{db: db, ticker: time.NewTicker(conf.keepaliveInterval), Conf: conf}

	util.Go(func() {
		for {
			if _, ok := <-msql.ticker.C; !ok {
				break
			}
			if err := db.DB().Ping(); err != nil {
				log.Debug("Mysql Ping: %v", err)
			}
		}
	})

	log.Info("msyql.New Connect To Mysql Success")

	return msql
}

type MgrConfig map[string][]Config

type MysqlMgr struct {
	instances map[string][]*Mysql
}

func (mgr *MysqlMgr) Get(tag string, args ...interface{}) *Mysql {
	pool, ok := mgr.instances[tag]
	if !ok {
		return nil
	}
	idx := uint64(0)
	if len(args) > 0 {
		if i, ok := args[0].(int); ok {
			idx = uint64(i)
		} else {
			idx = util.Hash(fmt.Sprintf("%v", args[0]))
		}
	}
	return pool[uint32(idx)%uint32(len(pool))]
}

func (mgr *MysqlMgr) ForEach(cb func(string, int, *Mysql)) {
	for tag, pool := range mgr.instances {
		for idx, m := range pool {
			cb(tag, idx, m)
		}
	}
}

func NewMgr(mgrConf MgrConfig) *MysqlMgr {
	mgr := &MysqlMgr{
		instances: map[string][]*Mysql{},
	}

	for tagstr, confs := range mgrConf {
		total := 0
		sort.Slice(confs, func(i, j int) bool {
			return confs[i].ID > confs[j].ID
		})
		for _, conf := range confs {
			instance := New(conf)
			tags := strings.Split(tagstr, ":")
			for _, tag := range tags {
				mgr.instances[tag] = append(mgr.instances[tag], instance)
			}
			total++
		}

		if total == 0 {
			panic("invalid MgrConfig, 0 config")
		}
	}

	return mgr
}

func ClearTransaction(tx *sql.Tx) error {
	err := tx.Rollback()
	if err != nil && err != sql.ErrTxDone {
		log.Error("ClearTransaction failed: %v\n", err)
	}
	return err
}
