package mgo

import (
	"time"

	"gopkg.in/mgo.v2"

	"github.com/liming8519/pholcus/common/pool"
	"github.com/liming8519/pholcus/config"
	"github.com/liming8519/pholcus/logs"
	"github.com/spf13/viper"
	"net"
	"crypto/tls"
)

type MgoSrc struct {
	*mgo.Session
}

var (
	connGcSecond = time.Duration(config.MGO_CONN_GC_SECOND) * 1e9
	session      *mgo.Session
	err          error
	MgoPool      = pool.ClassicPool(
		config.MGO_CONN_CAP,
		config.MGO_CONN_CAP/5,
		func() (pool.Src, error) {
			if err != nil || session.Ping() != nil {
				if session != nil {
					session.Close()
				}
				Refresh()
			}
			return &MgoSrc{session.Clone()}, err
		},
		connGcSecond)
)

func Refresh() {
	dialInfo:= &mgo.DialInfo{
			Addrs:    []string{viper.GetString("mongodb.address")},
			Timeout:  3 * time.Second,
			Database: viper.GetString("mongodb.database"),
			Username: viper.GetString("mongodb.username"),
			Password: viper.GetString("mongodb.password"),
			DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
				return tls.Dial("tcp", addr.String(), &tls.Config{})
			},
		}
	logs.Log.Error("MongoDB info：%v\n", dialInfo)
	var err error
	session, err = mgo.DialWithInfo(dialInfo)
	if err != nil {
		logs.Log.Error("MongoDB：%v\n", err)
	} else if err = session.Ping(); err != nil {
		logs.Log.Error("MongoDB：%v\n", err)
	} else {
		session.SetPoolLimit(config.MGO_CONN_CAP)
	}
}

// 判断资源是否可用
func (self *MgoSrc) Usable() bool {
	if self.Session == nil || self.Session.Ping() != nil {
		return false
	}
	return true
}

// 使用后的重置方法
func (*MgoSrc) Reset() {}

// 被资源池删除前的自毁方法
func (self *MgoSrc) Close() {
	if self.Session == nil {
		return
	}
	self.Session.Close()
}

func Error() error {
	return err
}

// 调用资源池中的资源
func Call(fn func(pool.Src) error) error {
	return MgoPool.Call(fn)
}

// 销毁资源池
func Close() {
	MgoPool.Close()
}

// 返回当前资源数量
func Len() int {
	return MgoPool.Len()
}

// 获取所有数据
func DatabaseNames() (names []string, err error) {
	err = MgoPool.Call(func(src pool.Src) error {
		names, err = src.(*MgoSrc).DatabaseNames()
		return err
	})
	return
}

// 获取数据库集合列表
func CollectionNames(dbname string) (names []string, err error) {
	MgoPool.Call(func(src pool.Src) error {
		names, err = src.(*MgoSrc).DB(dbname).CollectionNames()
		return err
	})
	return
}
