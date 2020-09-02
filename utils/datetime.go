package utils

import (
	"time"
)

// LocationCST 中国标准时间(CST)的location对象, 时区为+08:00
var LocationCST = time.FixedZone("CST", 8*60*60)
