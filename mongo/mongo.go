package mongo

import (
	"errors"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"gopkg.in/mgo.v2"
	"time"
)

//[mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]

const (
	defaultPoolSize          = 10
	defaultConcurrency       = 100
	defaultKeepaliveInterval = time.Second * 300
)

type Config struct {
	Addrs             []string  `json:"Addrs"`
	Username          string    `json:"Username"`
	Password          string    `json:"Password"`
	Database          string    `json:"Database"`
	PoolSize          int       `json:"PoolSize"`
	Concurrency       int       `json:"Concurrency"`
	Safe              *mgo.Safe `json:"Safe"`
	DialTimeout       int       `json:"DialTimeout"`
	SocketTimeout     int       `json:"SocketTimeout"`
	KeepaliveInterval int       `json:"KeepaliveInterval"`
	dialTimeout       time.Duration
	socketTimeout     time.Duration
	keepaliveInterval time.Duration
}

type MongoSessionWrap struct {
	*mgo.Session
	parent *Mongo
	tmp    *mgo.Session
}

func (wrap *MongoSessionWrap) init(conf Config) {
	if conf.Safe != nil {
		wrap.SetSafe(conf.Safe)
	}

	wrap.SetMode(mgo.Strong, true)
	wrap.SetPoolLimit(1)
	wrap.SetCursorTimeout(0)
	wrap.SetSocketTimeout(conf.socketTimeout)
}

func (wrap *MongoSessionWrap) Clone() *MongoSessionWrap {
	return &MongoSessionWrap{
		parent:  wrap.parent,
		tmp:     wrap.Session,
		Session: wrap.Session.Clone(),
	}
}

func (wrap *MongoSessionWrap) Close() {
	wrap.Session.Close()
	wrap.Session = wrap.tmp
	wrap.parent.chSession <- wrap
}

func (wrap *MongoSessionWrap) ping() error {
	sess := wrap.Session.Clone()
	defer sess.Close()
	return sess.Ping()
}

type Mongo struct {
	Conf      Config
	ticker    *time.Ticker
	chSession chan *MongoSessionWrap
	sessions  []*MongoSessionWrap
	chStop    chan util.Empty
}

func (m *Mongo) Session() *MongoSessionWrap {
	return <-m.chSession
}

func (m *Mongo) SessionWithTimeout(to time.Duration) (*MongoSessionWrap, error) {
	select {
	case sess := <-m.chSession:
		return sess, nil
	case <-time.After(to):
	}
	return nil, errors.New("mongodb get session timeout")
}

func (m *Mongo) Insert(db, collection string, doc interface{}) error {
	session := m.Session().Clone()
	defer session.Close()
	err := session.Session.DB(db).C(collection).Insert(doc)
	return err
}

func (m *Mongo) EnsureIndex(dbname string, cname string, keys []string) {
	index := mgo.Index{Key: keys}
	session := m.Session().Clone()
	defer session.Close()
	if err := session.DB(dbname).C(cname).EnsureIndex(index); err != nil {
		log.Info("mongodb %s.%s EnsureIndex Failed, Error: %v", dbname, cname, err)
	}
}

func (m *Mongo) Keepalive() {
	for {
		select {
		case <-m.chStop:
			for _, session := range m.sessions {
				session.Close()
			}
			return
		case <-m.ticker.C:
			for _, session := range m.sessions {
				util.Safe(func() {
					if err := session.ping(); err != nil {
						log.Debug("Mongo Ping: %v", err)
					}
				})
			}
		}
	}
}

func (m *Mongo) Stop() {
	m.ticker.Stop()
}

func New(conf Config) *Mongo {
	log.Info("mongo.New Connect To Mongo ...")
	if conf.DialTimeout > 0 {
		conf.dialTimeout = time.Second * time.Duration(conf.DialTimeout)
	} else {
		conf.dialTimeout = time.Second * 5
	}
	if conf.SocketTimeout > 0 {
		conf.socketTimeout = time.Second * time.Duration(conf.SocketTimeout)
	} else {
		conf.socketTimeout = time.Second * 5
	}
	if conf.KeepaliveInterval > 0 {
		conf.keepaliveInterval = time.Second * time.Duration(conf.KeepaliveInterval)
	} else {
		conf.keepaliveInterval = defaultKeepaliveInterval
	}
	if conf.PoolSize <= 0 {
		conf.PoolSize = defaultPoolSize
	}
	if conf.Concurrency <= 0 {
		conf.Concurrency = defaultConcurrency
	}

	dialInfo := mgo.DialInfo{
		Addrs:    conf.Addrs,
		Timeout:  conf.dialTimeout,
		Username: conf.Username,
		Password: conf.Password,
		Database: conf.Database,
	}

	mongo := &Mongo{Conf: conf, ticker: time.NewTicker(conf.keepaliveInterval), chSession: make(chan *MongoSessionWrap, conf.PoolSize*conf.Concurrency), chStop: make(chan util.Empty, 1)}

	for i := 0; i < conf.PoolSize; i++ {
		session, err := mgo.DialWithInfo(&dialInfo)
		//session, err := mgo.DialWithTimeout(conf.ConnString, conf.dialTimeout)
		if err != nil {
			log.Fatal("mongo.New mgo.DialWithTimeout failed: %v", err)
		}

		err = session.Ping()
		if err != nil {
			log.Fatal("mongo.New failed: %v", err)
		}

		mongo.sessions = append(mongo.sessions, &MongoSessionWrap{session, mongo, nil})
		mongo.sessions[i].init(conf)
		mongo.chSession <- mongo.sessions[i]
	}

	for i := 1; i < conf.Concurrency; i++ {
		for _, session := range mongo.sessions {
			mongo.chSession <- session
		}
	}

	util.Go(mongo.Keepalive)

	log.Info("mongo.New(pool size: %d) Connect To Mongo Success", len(mongo.sessions))

	return mongo
}
