package utils

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	logger "github.com/Sirupsen/logrus"
	uuid "github.com/nu7hatch/gouuid"
)

// HandlePanic 封装对 panic 的处理, 在开协程时比较有用
func HandlePanic(ctx context.Context, f func()) func() {
	reqID := RequestIDFromContext(ctx)
	return func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(reqID, "panic: ", err)
				stack := strings.Join(strings.Split(string(debug.Stack()), "\n")[2:], "\n")
				logger.Error(reqID, "stack: ", stack)
			}
		}()

		f()
	}
}

// HandlePanicV2 增加参数传递 避免开协程时的闭包问题
func HandlePanicV2(ctx context.Context, f func(interface{})) func(interface{}) {
	reqID := RequestIDFromContext(ctx)
	return func(arg interface{}) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(reqID, "panic: ", err)
				stack := strings.Join(strings.Split(string(debug.Stack()), "\n")[2:], "\n")
				logger.Error(reqID, "stack: ", stack)
			}
		}()

		f(arg)
	}
}

//RequestIDFromContext 从ctx中得到消息的RequestID
func RequestIDFromContext(ctx context.Context) string {
	reqID := requestIDFromContext(ctx)
	if reqID == "" {
		return GetRequestID()
	}
	return reqID
}

//GetRequestID generate a unique ID for the server
func GetRequestID() string {
	id, _ := uuid.NewV4()

	rid := id.String()
	if rid == "" || len(rid) < 8 { //取毫秒
		return fmt.Sprintf("%d", time.Now().UTC().UnixNano()/1000)
	}
	return rid[0:7] //取uuid的前8位
}

const (
	KeyRequestID = 1
)

//RequestIDFromContext 从ctx中得到消息的RequestID
func requestIDFromContext(ctx context.Context) string {
	id, ok := ctx.Value(KeyRequestID).(string)
	if ok {
		return id
	}
	return ""
}

//GetUUIDString 得到uuid
func GetUUIDString() string {
	id, _ := uuid.NewV4()
	ret := strings.Replace(id.String(), "-", "", -1)
	if ret == "" {
		ret = fmt.Sprintf("%d", time.Now().Nanosecond())
	}
	return ret
}
