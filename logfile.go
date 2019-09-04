package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"

	// "fmt"
	"os"

	"strings"
	//	"syscall"
	//	"time"
	"strconv"

	"encoding/json"

	"github.com/robfig/config"

	"errors"
	"flag"
	"reflect"
	"syscall"
)

//命令行参数配置
var (
	configFile  = flag.String("c", "file.ini", "config file")
	daemon      = flag.Bool("d", false, "run daemon")
	readlogpath = flag.String("log", "/var/log/nginx/access_json.log", "nginx log path")
	db_tmpdir   = flag.String("tmpdir", "/tmp", "db file path")
)

//配置项
type logConf struct {
	path   string
	tmpdir string
}

//nginx 日志interface
type NginxLog interface {
}

//变量类型
func IsMap(v interface{}) bool {
	typev := reflect.TypeOf(v).String()
	if string([]byte(typev)[:3]) == "map" {
		return true
	}
	return false
}

//获取配置项
func GetConf() logConf {
	//	conf := logConf{path: "/var/log/nginx/access_json.log", tmpdir: "/tmp"}
	c, err := config.ReadDefault(*configFile)
	var path, tmpdir string
	if err != nil {
		// fmt.Println("config file does not exsist")
		// panic(err)
		path = *readlogpath
		tmpdir = *db_tmpdir
	} else {
		path, _ = c.String("", "path")
		tmpdir, _ = c.String("", "tmpdir")
	}
	conf := logConf{path: path, tmpdir: tmpdir}
	return conf
}

// IsFile checks whether the path is a file,
// it returns false when it's a directory or does not exist.
func IsFile(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}

//读取文件，发送到队列
func ReadFile() (uint32, error) {
	conf := GetConf()
	//发送消息条数
	var num uint32 = 0
	//	file, err := os.Open(conf.path)
	file, err := os.Open(conf.path)
	if err != nil {
		// fmt.Println(err.Error())
		return num, err
	}
	// var int line := 0
	finfo, _ := file.Stat()
	stat_t := finfo.Sys().(*syscall.Stat_t)
	inoid := stat_t.Ino
	mtime := stat_t.Mtim.Sec
	line := 0
	tmpfilePath := strings.TrimRight(conf.tmpdir, "/") + "/" + md5V(conf.path) + ".tmp"
	// fmt.Println(tmpfilePath)
	isTmp := IsFile(tmpfilePath)
	var c_tmp *config.Config
	// var last_line int
	var last_inoid uint64
	var last_mtime int64
	if isTmp {
		c_tmp, err = config.ReadDefault(tmpfilePath)
		if err != nil {
			// fmt.Println(err.Error())
			return num, err
		}
		last_line, _ := c_tmp.Int("", "line")
		c_inoid, _ := c_tmp.Int("", "inoid")
		c_mtime, _ := c_tmp.Int("", "mtime")
		last_inoid = uint64(c_inoid)
		last_mtime = int64(c_mtime)
		// if err != nil {
		// 	fmt.Println(err.Error())
		// 	return num, err
		// }
		line = last_line + 1
	} else {
		c_tmp = config.NewDefault()
		last_inoid = 0
		last_mtime = 0
	}
	if last_mtime > mtime {
		// fmt.Println("file not change")
		return num, errors.New("file not change")
	}
	if last_inoid != inoid {
		line = 0
	}
	// line := 14
	i := 0
	txt := ""
	scanner := bufio.NewScanner(file)
	qq := DefaultProducer("logstash", "logstash")
	var m NginxLog
	// var bb []byte
	for scanner.Scan() {
		i++
		if i < line {
			continue
		}
		txt = scanner.Text()
		if txt != "" {
			err := json.Unmarshal([]byte(txt), &m)
			if err == nil && IsMap(m) {
				qq.MsgProducer(txt)
				num++
			}

		}
		// fmt.Println(txt)

	}
	// fmt.Println(i)
	c_tmp.AddOption("", "line", strconv.Itoa(i))
	c_tmp.AddOption("", "inoid", strconv.FormatUint(inoid, 10))
	c_tmp.AddOption("", "mtime", strconv.FormatInt(mtime, 10))
	c_tmp.WriteFile(tmpfilePath, 0644, "tmp file line")
	return num, err
}

//md5字符串
func md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
