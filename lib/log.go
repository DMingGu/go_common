package lib

import (
	log "github.com/sirupsen/logrus"
)

// 通用DLTag常量定义
const (
	DLTagUndefind     = "_undef"
	DLTagMySqlEmpty   = "_com_mysql_empty"
	DLTagMySqlFailed  = "_com_mysql_failure"
	DLTagMySqlSuccess = "_com_mysql_success"
	DLTagMySqlWarn    = "_com_mysqlwarn"
	DLTagRedisEmpty   = "_com_redis_empty"
	DLTagRedisSuccess = "_com_redis_success"
	DLTagRedisFailed  = "_com_redis_failure"
	DLTagHTTPSuccess  = "_com_http_success"
	DLTagHTTPFailed   = "_com_http_failure"
	DLTagTCPFailed    = "_com_tcp_failure"
	DLTagRequestIn    = "_com_request_in"
	DLTagRequestOut   = "_com_request_out"
	DLTagBaseFailed   = "_com_base_failure"
)
const (
	_dlTag          = "dltag"
	_traceId        = "traceid"
	_spanId         = "spanid"
	_dlTagBizPrefix = "_com_"
	_dlTagBizUndef  = "_com_undef"
)

var Log *Logger

type Trace struct {
	TraceId     string
	SpanId      string
	Caller      string
	SrcMethod   string
	HintCode    int64
	HintContent string
}

type TraceContext struct {
	Trace
}

type Logger struct {
}

func (l *Logger) TagInfo(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = dltag
	m[_traceId] = trace.TraceId
	m[_spanId] = trace.SpanId
	log.Info(m)
}

func (l *Logger) TagWarn(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = dltag
	m[_traceId] = trace.TraceId
	m[_spanId] = trace.SpanId
	log.Warn(m)
}

func (l *Logger) TagError(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = dltag
	m[_traceId] = trace.TraceId
	m[_spanId] = trace.SpanId
	log.Error(m)
}

func (l *Logger) TagTrace(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = dltag
	m[_traceId] = trace.TraceId
	m[_spanId] = trace.SpanId
	log.Trace(m)
}

func (l *Logger) TagDebug(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = dltag
	m[_traceId] = trace.TraceId
	m[_spanId] = trace.SpanId
	log.Debug(m)
}
