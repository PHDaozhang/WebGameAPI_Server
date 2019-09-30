package mulMongo

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

//var MongoClient *mgo.Session

var mongoClientMap map[string]*mgo.Session = map[string]*mgo.Session{}

func init() {}

// 初始化MongoDB
func InitMongo(mongoDbPre string) {
	db_host := beego.AppConfig.String( mongoDbPre +"db_host")
	auth_db := beego.AppConfig.String(mongoDbPre +"auth_db")
	auth_user := beego.AppConfig.String(mongoDbPre +"auth_user")
	auth_pass := beego.AppConfig.String(mongoDbPre +"auth_pass")
	pool_limit, _ := beego.AppConfig.Int(mongoDbPre +"pool_limit")

	dialInfo := &mgo.DialInfo{
		Addrs:     []string{db_host},
		Timeout:   60 * time.Second,
		Source:    auth_db,
		Username:  auth_user,
		Password:  auth_pass,
		PoolLimit: pool_limit,
	}

	s := &mgo.Session{}
	err := errors.New("")
	if auth_user != "" && auth_pass != "" {
		s, err = mgo.DialWithInfo(dialInfo)
	} else {
		s, err = mgo.Dial(db_host)
	}
	if err != nil {
		logs.Error(fmt.Sprintf("连接MongoDB失败: host:%s, %s\n", db_host, err))
	} else {
		logs.Trace("连接MongoDB成功 host:%s,db:%s",db_host,auth_db)
	}

	if err == nil {
		mongoClientMap[ auth_db ] = s
	}
	//MongoClient = s
}

func connect(db, collection string) (*mgo.Session, *mgo.Collection) {
	dbClient,ok := mongoClientMap[ db ]
	if !ok {
		return nil,nil
	}

	ms := dbClient.Copy()
	c := ms.DB(db).C(collection)
	ms.SetMode(mgo.Monotonic, true)
	return ms, c
}

func getDb(db string) (*mgo.Session, *mgo.Database) {
	dbClient,ok := mongoClientMap[ db ]
	if !ok {
		return nil,nil
	}

	ms := dbClient.Copy()
	return ms, ms.DB(db)
}

func IsEmpty(db, collection string) bool {
	ms, c := connect(db, collection)
	defer ms.Close()
	count, err := c.Count()
	if err != nil {
		log.Fatal(err)
	}
	return count == 0
}

func Count(db, collection string, query interface{}) (int, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	return c.Find(query).Count()
}

func Insert(db, collection string, docs ...interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Insert(docs...)
}

func FindOne(db, collection string, query, selector, result interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Find(query).Select(selector).One(result)
}

func FindAll(db, collection string, query, selector, result interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Find(query).Select(selector).All(result)
}

func FindPage(db, collection string, page, limit int, query, selector, result interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Find(query).Select(selector).Skip(page * limit).Limit(limit).All(result)
}

func FindIter(db, collection string, query interface{}) *mgo.Iter {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Find(query).Iter()
}

func Update(db, collection string, selector, update interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Update(selector, update)
}

func Upsert(db, collection string, selector, update interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	_, err := c.Upsert(selector, update)
	return err
}

func UpdateAll(db, collection string, selector, update interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	_, err := c.UpdateAll(selector, update)
	return err
}

func Remove(db, collection string, selector interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Remove(selector)
}

func RemoveAll(db, collection string, selector interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	_, err := c.RemoveAll(selector)
	return err
}

//insert one or multi documents
func BulkInsert(db, collection string, docs ...interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	bulk := c.Bulk()
	bulk.Insert(docs...)
	return bulk.Run()
}

func BulkRemove(db, collection string, selector ...interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(db, collection)
	defer ms.Close()

	bulk := c.Bulk()
	bulk.Remove(selector...)
	return bulk.Run()
}

func BulkRemoveAll(db, collection string, selector ...interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	bulk := c.Bulk()
	bulk.RemoveAll(selector...)
	return bulk.Run()
}

func BulkUpdate(db, collection string, pairs ...interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	bulk := c.Bulk()
	bulk.Update(pairs...)
	return bulk.Run()
}

func BulkUpdateAll(db, collection string, pairs ...interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	bulk := c.Bulk()
	bulk.UpdateAll(pairs...)
	return bulk.Run()
}

func BulkUpsert(db, collection string, pairs ...interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	bulk := c.Bulk()
	bulk.Upsert(pairs...)
	return bulk.Run()
}

func PipeAll(db, collection string, pipeline, result interface{}, allowDiskUse bool) error {
	ms, c := connect(db, collection)
	defer ms.Close()
	var pipe *mgo.Pipe
	if allowDiskUse {
		pipe = c.Pipe(pipeline).AllowDiskUse()
	} else {
		pipe = c.Pipe(pipeline)
	}
	return pipe.All(result)
}

func PipeOne(db, collection string, pipeline, result interface{}, allowDiskUse bool) error {
	ms, c := connect(db, collection)
	defer ms.Close()
	var pipe *mgo.Pipe
	if allowDiskUse {
		pipe = c.Pipe(pipeline).AllowDiskUse()
	} else {
		pipe = c.Pipe(pipeline)
	}
	return pipe.One(result)
}

func PipeIter(db, collection string, pipeline interface{}, allowDiskUse bool) *mgo.Iter {
	ms, c := connect(db, collection)
	defer ms.Close()
	var pipe *mgo.Pipe
	if allowDiskUse {
		pipe = c.Pipe(pipeline).AllowDiskUse()
	} else {
		pipe = c.Pipe(pipeline)
	}

	return pipe.Iter()

}

func Explain(db, collection string, pipeline, result interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()
	pipe := c.Pipe(pipeline)
	return pipe.Explain(result)
}
func GridFSCreate(db, prefix, name string) (*mgo.GridFile, error) {
	ms, d := getDb(db)
	defer ms.Close()
	gridFs := d.GridFS(prefix)
	return gridFs.Create(name)
}

func GridFSFindOne(db, prefix string, query, result interface{}) error {
	ms, d := getDb(db)
	defer ms.Close()
	gridFs := d.GridFS(prefix)
	return gridFs.Find(query).One(result)
}

func GridFSFindAll(db, prefix string, query, result interface{}) error {
	ms, d := getDb(db)
	defer ms.Close()
	gridFs := d.GridFS(prefix)
	return gridFs.Find(query).All(result)
}

func GridFSOpen(db, prefix, name string) (*mgo.GridFile, error) {
	ms, d := getDb(db)
	defer ms.Close()
	gridFs := d.GridFS(prefix)
	return gridFs.Open(name)
}

func GridFSRemove(db, prefix, name string) error {
	ms, d := getDb(db)
	defer ms.Close()
	gridFs := d.GridFS(prefix)
	return gridFs.Remove(name)
}

func FindOneSort(db, collection string, query, selector interface{}, sort string, result interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Find(query).Select(selector).Sort(sort).One(result)
}

func FindAllSort(db, collection string, query, selector interface{}, sort string, result interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Find(query).Select(selector).Sort(sort).All(result)
}

func FindAllByCondition(db, collection string, query, selector interface{}, sort string, skip, limit int, result interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()
	q := c.Find(query).Select(selector)
	if sort != "" {
		q.Sort(sort)
	}
	if skip != 0 {
		q.Skip(skip)
	}
	if limit != 0 {
		q.Limit(limit)
	}

	return q.All(result)
}

func RemoveById(db, collection string, id bson.ObjectId) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.RemoveId(id)
}