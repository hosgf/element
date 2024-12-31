package types

import (
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

func Parse(data string) (int64, string) {
	unit, _ := gregex.ReplaceString("[0-9]+", "", data)
	value := gconv.Int64(gstr.Replace(data, unit, ""))
	return value, unit
}
