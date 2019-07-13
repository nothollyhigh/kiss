package mongo

import (
	"errors"
	"github.com/nothollyhigh/kiss/log"
	"gopkg.in/mgo.v2"
	"time"
	//"github.com/nothollyhigh/kiss/util"
)

//[mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]

const defaultPoolSize = 10

type MongoConf struct {
	Addrs    []string `json:"Addrs"`
	Username string   `json:"Username"`
	Password string   `json:"Password"`
	Database string   `json:"Database"`
	//ConnString string `json:"ConnString"`
	PoolSize int `json:"PoolSize"`
	// PoolSizeMultiple int       `json:"PoolSizeMultiple"`
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

func (wrap *MongoSessionWrap) init(conf MongoConf) {
	if conf.Safe != nil {
		wrap.SetSafe(conf.Safe)
	}

	wrap.SetMode(mgo.Strong, true)
	wrap.SetSocketTimeout(conf.socketTimeout)
	wrap.SetCursorTimeout(0)
}

func (wrap *MongoSessionWrap) Clone() *MongoSessionWrap {
	wrap.tmp = wrap.Session
	wrap.Session = wrap.Session.Clone()
	return wrap
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
	Conf      MongoConf
	ticker    *time.Ticker
	chSession chan *MongoSessionWrap
	sessions  []*MongoSessionWrap
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

func NewMongo(conf MongoConf) *Mongo {
	log.Info("NewMongo Connect To Mongo ...")
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
		conf.keepaliveInterval = time.Second * 300
	}
	if conf.PoolSize <= 0 {
		conf.PoolSize = defaultPoolSize
	}

	dialInfo := mgo.DialInfo{
		Addrs:    conf.Addrs,
		Timeout:  conf.dialTimeout,
		Username: conf.Username,
		Password: conf.Password,
		Database: conf.Database,
	}

	session, err := mgo.DialWithInfo(&dialInfo)
	//session, err := mgo.DialWithTimeout(conf.ConnString, conf.dialTimeout)
	if err != nil {
		log.Fatal("NewMongo mgo.DialWithTimeout failed: %v", err)
	}

	err = session.Ping()
	if err != nil {
		log.Fatal("NewMongo failed: %v", err)
	}

	mongo := &Mongo{Conf: conf, ticker: time.NewTicker(conf.keepaliveInterval), chSession: make(chan *MongoSessionWrap, conf.PoolSize)}
	mongo.sessions = append(mongo.sessions, &MongoSessionWrap{session, mongo, nil})
	mongo.sessions[0].init(conf)
	mongo.chSession <- mongo.sessions[0]

	if conf.PoolSize > 1 {
		for i := 1; i < conf.PoolSize; i++ {
			sessionCopy := &MongoSessionWrap{session.Copy(), mongo, nil}
			sessionCopy.init(conf)
			mongo.chSession <- sessionCopy
			mongo.sessions = append(mongo.sessions, sessionCopy)
		}
	}

	// if conf.PoolSizeMultiple > 1 {
	// 	for i := 0; i < conf.PoolSizeMultiple-1; i++ {
	// 		for j := 0; j < conf.PoolSize; j++ {
	// 			sess := mongo.sessions[j]
	// 			mongo.sessions = append(mongo.sessions, &MongoSessionWrap{sess.Session, &sync.Mutex{}})
	// 		}
	// 	}
	// }

	// util.Go(func() {
	// 	for {
	// 		if _, ok := <-mongo.ticker.C; !ok {
	// 			break
	// 		}
	// 		for _, session := range mongo.sessions {
	// 			util.Safe(func() {
	// 				if err := session.ping(); err != nil {
	// 					log.Debug("Mongo Ping: %v", err)
	// 				}
	// 			})
	// 		}
	// 	}
	// })

	log.Info("NewMongo(pool size: %d) Connect To Mongo Success", len(mongo.sessions))

	return mongo
}
