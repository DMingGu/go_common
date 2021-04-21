package lib

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

var (
	TimeLocation *time.Location
	TimeFormat = "2006-01-02 15:04:05"
	DateFormat = "2006-01-02"
	YearMonthFormat = "200601"
	LocalIP = net.ParseIP("127.0.0.1")
)

//公共初始化函数：支持两种方式设置配置文件
func InitModule(configPath string, modules []string) error {
	conf := flag.String("config", configPath, "input config file like /config/dev/")
	flag.Parse()
	if *conf == "" {
		flag.Usage()
		os.Exit(1)
	}
	log.Println("----------------------------------start loading resources.--------------------------------------")
	Log.TagInfo(NewTrace(),"", map[string]interface{}{"config":*conf})
	// 设置ip信息，优先设置便于日志打印
	ips := GetLocalIPs()
	if len(ips) > 0 {
		LocalIP = ips[0]
	}
	// 解析配置文件目录
	if err := ParseConfPath(*conf); err != nil {
		return err
	}
	//初始化配置文件
	if err := InitViperConf(); err != nil {
		return err
	}
	initLog()//初始化日志
	// 加载base配置
	if InArrayString("base", modules) {
		InitBaseConf()//初始化base文件
	}

	// 加载redis配置
	if InArrayString("redis", modules) {
		if err := InitRedisPool(GetConfPath("redis_map")); err != nil {
			Log.TagError(NewTrace(),DLTagRedisFailed, map[string]interface{}{"err":err.Error()})
		}
	}

	// 加载mysql配置并初始化实例
	if InArrayString("mysql", modules) {
		if err := InitDBPool(GetConfPath("mysql_map")); err != nil {
			Log.TagError(NewTrace(),DLTagMySqlFailed, map[string]interface{}{"err":err.Error()})

		}
	}
	// 设置时区
	if location, err := time.LoadLocation(ConfBase.TimeLocation); err != nil {
		return err
	} else {
		TimeLocation = location
	}
	log.Println("----------------------------------success loading resources.--------------------------------------")
	return nil
}
func initLog()  {
	if GetConfEnv()=="dev" {
		// 设置日志格式为json格式
		log.SetFormatter(&log.TextFormatter{})
	}else{
		// 设置日志格式为json格式
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat:TimeFormat,
		})
	}
	// 设置将日志输出到标准输出（默认的输出为stderr，标准错误）
	// 日志消息输出可以是任意的io.writer类型
	log.SetOutput(os.Stdout)
	switch GetStringConf("base.log.log_level") {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.DebugLevel)

	}

}
//公共销毁函数
func Destroy() {

	log.Println("----------------------------------start destroy resources.---------------------------------------")
	//关闭mysql
	CloseDB()
	//关闭redis
	CloseRedis()
	log.Println("----------------------------------success destroy resources.--------------------------------------")
}

func HttpGET(trace *TraceContext, urlString string, urlParams url.Values, msTimeout int, header http.Header) (*http.Response, []byte, error) {
	startTime := time.Now().UnixNano()
	client := http.Client{
		Timeout: time.Duration(msTimeout) * time.Millisecond,
	}
	urlString = AddGetDataToUrl(urlString, urlParams)
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		log.Warn(map[string]interface{}{
			"dltag":DLTagHTTPFailed,
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "GET",
			"args":      urlParams,
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	if len(header) > 0 {
		req.Header = header
	}
	req = addTrace2Header(req, trace)
	resp, err := client.Do(req)
	if err != nil {
		log.Warn(map[string]interface{}{
			"dltag":DLTagHTTPFailed,
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "GET",
			"args":      urlParams,
			"err":       err.Error(),
		})
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "GET",
			"args":      urlParams,
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "GET",
			"args":      urlParams,
			"result":    Substr(string(body), 0, 1024),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	Log.TagInfo(trace, DLTagHTTPSuccess, map[string]interface{}{
		"url":       urlString,
		"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
		"method":    "GET",
		"args":      urlParams,
		"result":    Substr(string(body), 0, 1024),
	})
	return resp, body, nil
}

func HttpPOST(trace *TraceContext, urlString string, urlParams url.Values, msTimeout int, header http.Header, contextType string) (*http.Response, []byte, error) {
	startTime := time.Now().UnixNano()
	client := http.Client{
		Timeout: time.Duration(msTimeout) * time.Millisecond,
	}
	if contextType == "" {
		contextType = "application/x-www-form-urlencoded"
	}
	urlParamEncode := urlParams.Encode()
	req, err := http.NewRequest("POST", urlString, strings.NewReader(urlParamEncode))
	if len(header) > 0 {
		req.Header = header
	}
	req = addTrace2Header(req, trace)
	req.Header.Set("Content-Type", contextType)
	resp, err := client.Do(req)
	if err != nil {
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "POST",
			"args":      Substr(urlParamEncode, 0, 1024),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "POST",
			"args":      Substr(urlParamEncode, 0, 1024),
			"result":    Substr(string(body), 0, 1024),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	Log.TagInfo(trace, DLTagHTTPSuccess, map[string]interface{}{
		"url":       urlString,
		"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
		"method":    "POST",
		"args":      Substr(urlParamEncode, 0, 1024),
		"result":    Substr(string(body), 0, 1024),
	})
	return resp, body, nil
}

func HttpJSON(trace *TraceContext, urlString string, jsonContent string, msTimeout int, header http.Header) (*http.Response, []byte, error) {
	startTime := time.Now().UnixNano()
	client := http.Client{
		Timeout: time.Duration(msTimeout) * time.Millisecond,
	}
	req, err := http.NewRequest("POST", urlString, strings.NewReader(jsonContent))
	if len(header) > 0 {
		req.Header = header
	}
	req = addTrace2Header(req, trace)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "POST",
			"args":      Substr(jsonContent, 0, 1024),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Log.TagWarn(trace, DLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "POST",
			"args":      Substr(jsonContent, 0, 1024),
			"result":    Substr(string(body), 0, 1024),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	Log.TagInfo(trace, DLTagHTTPSuccess, map[string]interface{}{
		"url":       urlString,
		"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
		"method":    "POST",
		"args":      Substr(jsonContent, 0, 1024),
		"result":    Substr(string(body), 0, 1024),
	})
	return resp, body, nil
}

func AddGetDataToUrl(urlString string, data url.Values) string {
	if strings.Contains(urlString, "?") {
		urlString = urlString + "&"
	} else {
		urlString = urlString + "?"
	}
	return fmt.Sprintf("%s%s", urlString, data.Encode())
}

func addTrace2Header(request *http.Request, trace *TraceContext) *http.Request {
	traceId := trace.TraceId
	spanId := trace.SpanId
	if traceId != "" {
		request.Header.Set("didi-header-rid", traceId)
	}
	if spanId != "" {
		request.Header.Set("didi-header-spanid", spanId)
	}
	return request
}
/**
 * @author DMing
 * @description //TODO 获取md5
 * @date 10:43 上午 2019/2/3
 * @param
 * @return
 **/
func DMGetMd5Hash(data string) (string, error) {
	h := md5.New()
	_, err := h.Write([]byte(data))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
/**
 * @author DMing
 * @description //TODO 获取年月
 * @date 10:43 上午 2019/2/3
 * @param
 * @return
 **/
func DmGetYearMonth(time time.Time) string {
	timeStr := time.Format(YearMonthFormat)
	return timeStr
}

func NewTrace() *TraceContext {
	trace := &TraceContext{}
	trace.TraceId = GetTraceId()
	trace.SpanId = NewSpanId()
	return trace
}

func NewSpanId() string {
	timestamp := uint32(time.Now().Unix())
	ipToLong := binary.BigEndian.Uint32(LocalIP.To4())
	b := bytes.Buffer{}
	b.WriteString(fmt.Sprintf("%08x", ipToLong^timestamp))
	b.WriteString(fmt.Sprintf("%08x", rand.Int31()))
	return b.String()
}

func GetTraceId() (traceId string) {
	return calcTraceId(LocalIP.String())
}

func calcTraceId(ip string) (traceId string) {
	now := time.Now()
	timestamp := uint32(now.Unix())
	timeNano := now.UnixNano()
	pid := os.Getpid()

	b := bytes.Buffer{}
	netIP := net.ParseIP(ip)
	if netIP == nil {
		b.WriteString("00000000")
	} else {
		b.WriteString(hex.EncodeToString(netIP.To4()))
	}
	b.WriteString(fmt.Sprintf("%08x", timestamp&0xffffffff))
	b.WriteString(fmt.Sprintf("%04x", timeNano&0xffff))
	b.WriteString(fmt.Sprintf("%04x", pid&0xffff))
	b.WriteString(fmt.Sprintf("%06x", rand.Int31n(1<<24)))
	b.WriteString("b0") // 末两位标记来源,b0为go

	return b.String()
}

func GetLocalIPs() (ips []net.IP) {
	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, address := range interfaceAddr {
		ipNet, isValidIpNet := address.(*net.IPNet)
		if isValidIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP)
			}
		}
	}
	return ips
}

func InArrayString(s string, arr []string) bool {
	sort.Strings(arr)
	index := sort.SearchStrings(arr, s)
	if index < len(arr) && arr[index] == s {
		return true
	}
	return false
}

//Substr 字符串的截取
func Substr(str string, start int64, end int64) string {
	length := int64(len(str))
	if start < 0 || start > length {
		return ""
	}
	if end < 0 {
		return ""
	}
	if end > length {
		end = length
	}
	return string(str[start:end])
}
