package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *File  `json:"data"`
}

type File struct {
	RemotePath string `json:"remotePath"`
	Domain     string `json:"domain"`
	Path       string `json:"path"`
	NewName    string `json:"newName"`
	OriginName string `json:"originName"`
	Size       int64  `json:"size"`
}

var h bool
var datadir string
var domain string
var port int
var group string

func init() {
	flag.BoolVar(&h, "h", false, "帮助")
	flag.StringVar(&datadir, "D", "data", "设置数据存储路径")
	flag.StringVar(&domain, "d", "http://127.0.0.1:8080", "设置域名")
	flag.IntVar(&port, "p", 8080, "设置端口")
	flag.StringVar(&group, "g", "", "设置组名")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintln(os.Stderr, "Godfs version: Godfs/1.0.0")
	fmt.Fprintln(os.Stderr, "Usage: Godfs [-h 帮助] [-d 设置域名] [-D 设置数据存储路径] [-p 设置端口] [-g 设置组名]")
	fmt.Fprintln(os.Stderr, "启动成功后,打开浏览器访问端口即可，例如:http://127.0.0.1:8080")
	fmt.Fprintln(os.Stderr, "Options:")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if h {
		flag.Usage()
		return
	}

	fmt.Println("----------------------------------------------------------------")
	fmt.Println(fmt.Sprintf("\t\t启动端口: %d", port))
	fmt.Println(fmt.Sprintf("\t\t存储路径: %s", datadir))
	fmt.Println(fmt.Sprintf("\t\t访问域名: %s", domain))
	fmt.Println(fmt.Sprintf("\t\t访问组名: %s", group))
	fmt.Println(fmt.Sprintf("\t\t上传路径: /upload"))
	fmt.Println("----------------------------------------------------------------")

	initLog()

	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/", handleDownload)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func initLog() {
	// 创建日志目录
	err := os.MkdirAll("./logs", os.ModePerm)
	if err != nil {
		log.Println("create dir [./logs] failed!")
		return
	}
	// 创建日志文件
	datePath := time.Now().Format("2006.01.02")
	file, err := os.OpenFile(fmt.Sprintf("./logs/server.%s.log", datePath), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) // 打开日志文件，不存在则创建
	if err != nil {
		log.Println("create file [./logs/server.log] failed!")
		return
	}

	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime)
}

func fail(message string, w http.ResponseWriter) {
	result := &Result{}
	result.Code = http.StatusBadRequest
	result.Message = message
	bytes, _ := json.Marshal(result)
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(bytes)
}

func success(data *File, w http.ResponseWriter) {
	result := &Result{}
	result.Code = http.StatusOK
	result.Message = "Upload success"
	result.Data = data
	bytes, _ := json.Marshal(result)
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(bytes)
}

// 上传文件
func handleUpload(w http.ResponseWriter, request *http.Request) {
	// 允许跨域
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if request.Method == http.MethodOptions {
		return
	}

	// 文件上传只允许POST
	if request.Method != http.MethodPost {
		fail("Method not allowed", w)
		return
	}

	log.Println("----------文件上传开始----------")
	// 从表单中读取文件
	file, fileHeader, err := request.FormFile("file")
	if err != nil {
		fail("Read file error", w)
		return
	}
	defer file.Close()

	log.Println(fmt.Sprintf("源文件:%s", fileHeader.Filename))

	// 获取后缀
	ext := ""
	index := strings.LastIndex(fileHeader.Filename, ".")
	if index != -1 {
		ext = fileHeader.Filename[index+1:]
	}
	log.Println(fmt.Sprintf("后缀:%s", ext))

	// 新文件名
	now := time.Now()
	filename := fmt.Sprintf("%d-%s.%s", now.UnixNano()/1e6, RandString(6), ext)
	log.Println(fmt.Sprintf("filename:%s", filename))

	// 文件存储路径
	datePath := now.Format("/2006/01/02/")
	err = os.MkdirAll(fmt.Sprintf("%s%s", datadir, datePath), os.ModePerm)
	if err != nil {
		fail("Create file dir error", w)
		return
	}

	// 创建新文件
	newFile, err := os.Create(fmt.Sprintf("%s%s%s", datadir, datePath, filename))
	if err != nil {
		fail("Create file error", w)
		return
	}
	defer newFile.Close()

	// 将源文件写入新文件
	_, err = io.Copy(newFile, file)
	if err != nil {
		fail("Write file error", w)
		return
	}

	log.Println("----------文件上传结束----------")

	// 返回数据
	data := &File{}
	data.Domain = domain

	url := domain
	if group != "" {
		url = "/" + group
	}
	data.Path = fmt.Sprintf("%s%s", datePath, filename)
	data.RemotePath = url + data.Path
	data.OriginName = fileHeader.Filename
	data.NewName = filename
	data.Size = fileHeader.Size
	success(data, w)
}

// 下载
func handleDownload(w http.ResponseWriter, request *http.Request) {
	// 允许跨域
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if request.Method == http.MethodOptions {
		return
	}
	log.Println(fmt.Sprintf("请求路径:%s", request.URL.Path))

	// 文件下载只允许GET
	if request.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fail("Method not allowed", w)
		return
	}

	log.Println("----------文件下载开始----------")

	// URL转换为本地路径
	fullPath := datadir + request.URL.Path

	// 文件名
	index := strings.LastIndex(fullPath, "/")
	filename := fullPath[index+1:]
	log.Println(fmt.Sprintf("请求文件名:%s", filename))

	// 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Not Found")
		return
	}
	defer file.Close()

	log.Println("----------文件下载结束----------")

	// 设置响应的header头
	w.Header().Add("Content-type", "application/octet-stream")
	w.Header().Add("content-disposition", "attachment; filename=\""+filename+"\"")
	// 将文件写至responseBody
	_, err = io.Copy(w, file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Bad request")
		return
	}
}

// 随机字符串
func RandString(len int) string {
	base := "0123456789abcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano() / 1e6))
	result := ""
	for i := 0; i < len; i++ {
		start := r.Intn(46)
		result = result + base[start:start+1]
	}
	return result
}

func InitConfig(path string) map[string]string {
	// 初始化
	confMap := make(map[string]string)

	// 打开文件指定目录，返回一个文件f和错误信息
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// 创建一个输出流向该文件的缓冲流*Reader
	r := bufio.NewReader(f)
	for {
		// 读取，返回[]byte 单行切片给b
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		// 去除单行属性两端的空格
		s := strings.TrimSpace(string(b))
		// fmt.Println(s)

		// 判断等号=在该行的位置
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}
		// 取得等号左边的key值，判断是否为空
		key := strings.TrimSpace(s[:index])
		if len(key) == 0 {
			continue
		}

		// 取得等号右边的value值，判断是否为空
		value := strings.TrimSpace(s[index+1:])
		if len(value) == 0 {
			continue
		}
		// 这样就成功吧配置文件里的属性key=value对，成功载入到内存中c对象里
		confMap[key] = value
	}
	return confMap
}
