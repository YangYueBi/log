package log

import (
	"encoding/xml"
	"fmt"
	"github.com/cihub/seelog"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//----------------------------日志的参数------------------------------------------------------
type LogConst struct {
	ServerID       string    `服务id，初始化时传入的国标id`
	ProcessID      string    `过程id，数据外壳中包含`
	ProcessType    string    `过程类型，文档中`
	AcessType      string    `接入类型，目前未知`
	DataType       string    `数据类型，数据外壳中`
	VendorID       string    `厂商ID，目前未知`
	DeviceID       string    `设备ID，目前未知`
	SourceDataID   string    `源数据ID，目前未知`
	DestineDataID  string    `目的数据id，目前未知`
	SourceDataTime time.Time `源数据时间`
	ExceptionType  string    `异常类型`
}

const (
	//数据类型
	MotorVehicle, Face, Person, NonMotorVehicle string = "01", "02", "03", "04"
	//操作码
	Stop, Start, Regiest, UnResiest, Subscribtion, Disposition, Query, Configuration string = "100", "101", "102", "103", "104", "105", "106", "107"
	//状态吗
	Normal, Exception, Online, Offline string = "200", "201", "202", "203"
	//过程类型
	Transform, DataClean, DicConversion, StrConversion, Acess, OutPut string = "1001", "1002", "1003", "1004", "1005", "1006"
	//接入类型
	GAT1400Acess, HikiSdk, DahuaSdk, OrcAcess, MySqlAcess, FtpAcess, KafkaAcess string = "1101", "1102", "1103", "1104", "1105", "1106", "1107"
	//异常类型
	TransFailed, DataCleanFailed, DicConvtFailed, StrConvtFailed, KafkaProFailed, LoadFailed, OutputFailed string = "1201", "1202", "1203", "1204", "1205", "1206", "1207"
)

type StruLog struct {
	OperationFileName, StatusFileName, DataFileName, ErrorDataFileName, DebugFileName string
}

type Logger struct {
	//param     StruLog
	Operation *MyLog
	Status    *MyLog
	Data      *MyLog
	ErrorData *MyLog
	Debug     *MyLog
}

func NewLogger(param StruLog) *Logger {
	var operate, status, data, error, debug *MyLog = nil, nil, nil, nil, nil
	if param.OperationFileName != "" {
		operate = NewLog()
		if !operate.Init(getCurrentName() + "/" + param.OperationFileName) {
			fmt.Println("初始化操作日志对象失败")
			return nil
		}
	} else {
		fmt.Println("初始化操作日志对象成功")
	}
	if param.StatusFileName != "" {
		status = NewLog()
		if !status.Init(getCurrentName() + "/" + param.StatusFileName) {
			fmt.Println("初始化状态日志对象失败")
			return nil
		}
	} else {
		fmt.Println("初始化状态日志对象成功")
	}
	if param.DataFileName != "" {
		data = NewLog()
		if !data.Init(getCurrentName() + "/" + param.DataFileName) {
			fmt.Println("初始化数据日志对象失败")
			return nil
		}
	} else {
		fmt.Println("初始化数据日志对象成功")
	}
	if param.ErrorDataFileName != "" {
		error = NewLog()
		if !error.Init(getCurrentName() + "/" + param.ErrorDataFileName) {
			fmt.Println("初始化错误日志对象失败")
			return nil
		}
	} else {
		fmt.Println("初始化错误日志对象成功")
	}
	if param.DebugFileName != "" {
		debug = NewLog()
		if !debug.Init(getCurrentName() + "/" + param.DebugFileName) {
			fmt.Println("初始化Debug日志对象失败")
			return nil
		}
	} else {
		fmt.Println("初始化Debug日志对象成功")
	}

	//所有的日志对象完成初始化，启动修改日志对象的接口
	go startListen()
	return &Logger{operate, status, data, error, debug}
}
func (this *Logger) Flush() {
	this.Operation.Flush()
	this.Status.Flush()
	this.Data.Flush()
	this.ErrorData.Flush()
	this.Debug.Flush()
}

func startListen() bool {
	http.HandleFunc("/alter", alterLevel)
	addr := "0.0.0.0:44444"
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Rest监听 %s 失败，错误：%v", addr, err)
		return false
	} else {
		fmt.Printf("Rest监听 %s 成功", addr)
		return true
	}
}

func alterLevel(w http.ResponseWriter, r *http.Request) {

}

//===============================================================

type MyLog struct {
	bInit bool                   //是否初始化
	log   seelog.LoggerInterface //seelog对象
}

func NewLog() *MyLog {
	return &MyLog{bInit: false, log: nil}
}

func (this *MyLog) Init(path string) bool {
	if this.bInit == true {
		fmt.Println("对象已经初始化过")
		return false
	}

	//读取环境变量下的配置文件  D:\ICV_ROOT
	icvRoot := os.Getenv("ICV_ROOT")
	if icvRoot == "" {
		fmt.Println("获取ICV_ROOT的环境变量失败")
		return false
	}
	fmt.Println("获取ICV_ROOT的环境变量成功")

	icvRoot = strings.Replace(icvRoot, "\\", "/", -1)
	//替换//为/
	path = strings.Replace(path, "\\", "/", -1)
	//读取配置文件
	logCfgPath := icvRoot + "/Config/GoLogConfig.xml"
	var ok bool = false
	cfg, ok := readCfg(logCfgPath)
	if !ok {
		fmt.Println("读取配置文件失败")
		return false
	}
	fmt.Println("读取配置文件成功")

	//创建GoLog文件夹
	goLog := icvRoot + "/GoLog"
	if err := os.MkdirAll(goLog, os.ModePerm); err != nil {
		fmt.Println("创建goLog文件夹失败")
		return false
	}
	//日志的存储目录(文件夹)
	tmpFolder := ""
	allFolder := ""
	if path != "" {
		tmpFolder = path
	} else {
		tmpFolder = cfg.LogCfg.Filename
	}
	folders := strings.Split(tmpFolder, "/")
	for _, folder := range folders {
		if folder != "" && -1 == strings.Index(folder, ".") {
			allFolder = allFolder + "/" + folder
		}
	}
	if err := os.MkdirAll(goLog+allFolder, 0711); err != nil {
		fmt.Printf("创建文件夹%s失败", goLog+allFolder)
		return false
	}

	logFolder := goLog + "/" + tmpFolder
	level := 7
	switch strings.ToUpper(cfg.LogCfg.Loglevel) {
	case "DEBUG":
		{
			level = 7
		}
	case "INFO":
		{
			level = 6
		}
	case "NOTICE":
		{
			level = 5
		}
	case "WARN":
		{
			level = 4
		}
	case "ERROR":
		{
			level = 3

		}
	case "CRITICAL":
		{
			level = 2
		}
	case "ALERT":
		{
			level = 1
		}
	case "OFF":
		{
			level = 0
		}
	default:
		level = 7
	}

	//生成临时的xml配置文件
	maxrolls, _ := strconv.Atoi(cfg.LogCfg.Maxrolls)
	maxsize, _ := strconv.Atoi(cfg.LogCfg.Maxsize)
	bConsole := false
	if strings.ToLower(cfg.LogCfg.Console) == "true" {
		bConsole = true
	} else {
		bConsole = false
	}
	cfgByte := generateCfgXml(cfgStr{Filename: logFolder, Maxrolls: maxrolls, Maxsize: maxsize, Level: level}, bConsole)
	if cfgByte != nil {
		var err error
		fmt.Printf("配置信息是：%s", string(cfgByte))
		this.log, err = seelog.LoggerFromConfigAsBytes(cfgByte)
		if err != nil {
			this.bInit = false
			fmt.Println("解析临时配置文件失败，错误: ", err)
			return false
		}
		this.log.SetAdditionalStackDepth(1)
		this.bInit = true

		fmt.Println("初始化日志对象成功")

	} else {
		fmt.Println("生成配置临时文件失败")
		return false
	}
	return true
}
func (this *MyLog) Flush() {
	if this == nil {
		return
	}
	this.log.Flush()
}

func (this *MyLog) Trace(format string, params ...interface{}) {
	if this == nil {
		panic("log object not initialization")
	}
	this.log.Tracef(format, params...)
}
func (this *MyLog) Debug(format string, params ...interface{}) {
	if this == nil {
		panic("log object not initialization")
	}
	this.log.Debugf(format, params...)
}
func (this *MyLog) Notice(format string, params ...interface{}) {
	if this == nil {
		panic("log object not initialization")
	}
	this.log.Infof(format, params...)
}
func (this *MyLog) Info(format string, params ...interface{}) {
	if this == nil {
		panic("log object not initialization")
	}
	this.log.Infof(format, params...)
}
func (this *MyLog) Warn(format string, params ...interface{}) {
	if this == nil {
		panic("log object not initialization")
	}
	this.log.Warnf(format, params...)
}
func (this *MyLog) Error(format string, params ...interface{}) {
	if this == nil {
		panic("log object not initialization")
	}
	this.log.Errorf(format, params...)
}

func (this *MyLog) Critical(format string, params ...interface{}) {
	if this == nil {
		panic("log object not initialization")
	}
	this.log.Criticalf(format, params...)
}
func (this *MyLog) Emergency(format string, params ...interface{}) {
	if this == nil {
		panic("log object not initialization")
	}
	this.log.Criticalf(format, params...)
}

//===============================================================
//定义父结构体
type goLogCfg struct {
	XMLName xml.Name `xml:"LogConf"`
	LogCfg  Config   `xml:"Config"`
}
type Config struct {
	Console  string `xml:"console,attr"`
	Filename string `xml:"filename,attr"`
	Loglevel string `xml:"level,attr"`
	Maxlines string `xml:"maxlines,attr"`
	Maxsize  string `xml:"maxsize,attr"`
	Maxrolls string `xml:"maxrolls,attr"`
	Color    string `xml:"color,attr"`
}

func readCfg(cfgPath string) (goLogCfg, bool) {
	ret := goLogCfg{}
	file, err := os.Open(cfgPath)
	if err != nil {
		fmt.Println("读取文件异常,异常信息为:>", err)
		return ret, false
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("解析json并绑定至对象异常,异常信息为:", err)
		return ret, false
	}
	//fmt.Println(string(data))
	err = xml.Unmarshal(data, &ret)
	if err != nil {
		fmt.Println("反序列化失败，错误：", err)
		return ret, false
	}
	return ret, true
}
func writeTmpCfg(content []byte, tmpPath string) bool {
	if ioutil.WriteFile(tmpPath, content, 0644) != nil {
		return false
	}
	return true
}

//===============================================================
type seelogCfg struct {
	XMLName xml.Name   `xml:"seelog"`
	Levels  string     `xml:"levels,attr"`
	Outputs tagOutputs `xml:"outputs"`
	Formats tagFormats `xml:"formats"`
}
type seelogCfgNonConsole struct {
	XMLName xml.Name             `xml:"seelog"`
	Levels  string               `xml:"levels,attr"`
	Outputs tagOutputsNonConsole `xml:"outputs"`
	Formats tagFormats           `xml:"formats"`
}
type tagOutputs struct {
	TagFormatid string      `xml:"formatid,attr"`
	Console     console     `xml:"console"`
	Rollingfile rollingfile `xml:"rollingfile"`
}
type tagOutputsNonConsole struct {
	TagFormatid string      `xml:"formatid,attr"`
	Rollingfile rollingfile `xml:"rollingfile"`
}

type console struct {
}
type rollingfile struct {
	Formatid string `xml:"formatid,attr"`
	Type     string `xml:"type,attr"`
	Filename string `xml:"filename,attr"`
	Maxsize  string `xml:"maxsize,attr"`
	Maxrolls string `xml:"maxrolls,attr"`
}
type tagFormats struct {
	Formats format `xml:"format"`
}
type format struct {
	Id     string `xml:"id,attr"`
	Format string `xml:"format,attr"`
}

type cfgStr struct {
	Filename string //日志的文件存储名成
	Maxsize  int    //单个文件的大小，b
	Maxrolls int    //最多几个文件
	Level    int    //日志级别
}

func generateCfgXml(cfgInfo cfgStr, bConsole bool) []byte {
	strLev := ""
	switch cfgInfo.Level {
	case 0:
		strLev = "off"
	case 1: //alert
		strLev = "critical"
	case 2: //critical
		strLev = "critical"
	case 3: //error
		strLev = "error,critical"
	case 4: //warn
		strLev = "warn,error,critical"
	case 5: //notice
		strLev = "info,warn,error,critical"
	case 6: //info
		strLev = "info,warn,error,critical"
	case 7: //debug
		strLev = "trace,debug,info,warn,error,critical"
	}
	if bConsole {
		cfg := seelogCfg{}
		cfg.Levels = strLev
		cfg.Outputs.TagFormatid = "main"

		cfg.Outputs.Rollingfile.Formatid = "main"
		cfg.Outputs.Rollingfile.Type = "size"
		cfg.Outputs.Rollingfile.Filename = cfgInfo.Filename
		cfg.Outputs.Rollingfile.Maxsize = strconv.Itoa(cfgInfo.Maxsize)
		cfg.Outputs.Rollingfile.Maxrolls = strconv.Itoa(cfgInfo.Maxrolls)

		cfg.Formats.Formats.Format = "%Date(2006/01/02 15:04:05) [%LEVEL] [%File:%Line] %Msg%n"
		cfg.Formats.Formats.Id = "main"
		output, err := xml.MarshalIndent(&cfg, "", "\t")
		if err != nil {
			fmt.Println("写xml失败，错误：", err)
			return nil
		}
		return output
	} else {
		cfg := seelogCfgNonConsole{}
		cfg.Levels = strLev
		cfg.Outputs.TagFormatid = "main"

		cfg.Outputs.Rollingfile.Formatid = "main"
		cfg.Outputs.Rollingfile.Type = "size"
		cfg.Outputs.Rollingfile.Filename = cfgInfo.Filename
		cfg.Outputs.Rollingfile.Maxsize = strconv.Itoa(cfgInfo.Maxsize)
		cfg.Outputs.Rollingfile.Maxrolls = strconv.Itoa(cfgInfo.Maxrolls)

		cfg.Formats.Formats.Format = "%Date(2006/01/02 15:04:05) [%LEVEL] [%File:%Line] %Msg%n"
		cfg.Formats.Formats.Id = "main"
		output, err := xml.MarshalIndent(&cfg, "", "\t")
		if err != nil {
			fmt.Println("写xml失败，错误：", err)
			return nil
		}
		return output
	}
}

//===============================================================
func getCurrentDirectory() string {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return ""
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return ""
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return ""
	}
	return string(path[0 : i+1])
}

func getCurrentName() string {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return ""
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return ""
	}
	return strings.Split(filepath.Base(path), ".")[0]
}
