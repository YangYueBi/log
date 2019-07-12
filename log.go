package log

import (
	"encoding/xml"
	"fmt"
	"github.com/cihub/seelog"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

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
		return true
	}

	//获取当前目录
	proName := getFullPath()
	if proName == "" {
		fmt.Println("获取当前目录失败")
		return false
	}
	//读取配置文件
	logCfgPath := proName + "/logcfg.xml"
	var ok bool = false
	cfg, ok := readCfg(logCfgPath)
	if !ok {
		fmt.Println("读取配置文件失败")
		return false
	}
	fmt.Println("读取配置文件成功")

	//创建GoLog文件夹
	goLog := proName + "/log"
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
	if err := os.MkdirAll(goLog+allFolder, 0666); err != nil {
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
		//fmt.Printf("配置信息是：%s", string(cfgByte))
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
func (this MyLog) Printf(format string, params ...interface{}) {
	this.log.Warnf(format, params...)
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

func getFullPath() string {
	exec, _ := filepath.Abs(filepath.Base(os.Args[0]))
	switch strOs := runtime.GOOS; strOs {
	case "windows":
		{
			pos := strings.LastIndex(exec, "\\")
			return exec[0:pos]
		}
	case "linux":
		{
			pos := strings.LastIndex(exec, "/")
			return exec[0:pos]
		}
	default:
		return ""
	}
}
