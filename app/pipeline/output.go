package pipeline

import (
	"sort"

	"github.com/liming8519/pholcus/app/pipeline/collector"
	"github.com/liming8519/pholcus/common/kafka"
	"github.com/liming8519/pholcus/common/mgo"
	"github.com/liming8519/pholcus/common/mysql"
	"github.com/liming8519/pholcus/runtime/cache"
)

// 初始化输出方式列表collector.DataOutputLib
func init() {
	for out, _ := range collector.DataOutput {
		collector.DataOutputLib = append(collector.DataOutputLib, out)
	}
	sort.Strings(collector.DataOutputLib)
}

// 刷新输出方式的状态
func RefreshOutput() {
	switch cache.Task.OutType {
	case "mgo":
		mgo.Refresh()
	case "mysql":
		mysql.Refresh()
	case "kafka":
		kafka.Refresh()
	}
}
