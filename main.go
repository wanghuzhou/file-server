package main

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/jinzhu/configor"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var (
	BaseFilePath string
	ServerPort   string
	db           *sql.DB
)

func main() {

	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/download", handleDownload)

	err := http.ListenAndServe(ServerPort, nil)
	if err != nil {
		log.Fatal("Server run fail")
	}
}

// 1.配置文件初始化
// 2.数据库初始化
func init() {
	// 配置文件
	var conf = Config{}
	err := configor.Load(&conf, "config.yml")
	if err != nil {
		panic(err)
	}
	BaseFilePath = conf.Server.BaseFilePath
	ServerPort = ":" + conf.Server.Port

	fmt.Printf("%v \n", conf)
	// 数据库
	//connStr := "postgres://postgres:postgres@192.168.122.11/file-server?sslmode=verify-full"
	connStr := "postgres://postgres:postgres@192.168.122.11/file-server?sslmode=disable"
	db, err = sql.Open("postgres", connStr)

	id := 4
	var hash string
	err = db.QueryRow("SELECT hash FROM file where id=$1", id).Scan(&hash)

	var f FileEntity
	err = db.QueryRow("SELECT * FROM file where id=$1", id).Scan(&f)

	fmt.Println(hash)
	fmt.Println(f)

}

func handleUpload(w http.ResponseWriter, request *http.Request) {

	fmt.Println("path", request.URL.Path)
	fmt.Println("scheme", request.URL.Scheme)
	fmt.Println(request.Form["url_long"])

	//文件上传只允许POST方法
	if request.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("Method not allowed"))
		return
	}

	//从表单中读取文件
	file, fileHeader, err := request.FormFile("file")
	if err != nil {
		_, _ = io.WriteString(w, "Read file error")
		return
	}
	//defer 结束时关闭文件
	defer file.Close()
	log.Println("filename: " + fileHeader.Filename)

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		_, _ = io.WriteString(w, "Upload fail")
	}

	filebytes := buf.Bytes()
	md5str := md5Str(filebytes)

	//创建文件
	newFile, err := os.Create(BaseFilePath + "/" + md5str + fileHeader.Filename)
	if err != nil {
		_, _ = io.WriteString(w, "Create file error")
		return
	}
	//defer 结束时关闭文件
	defer newFile.Close()

	_, err = newFile.Write(filebytes)
	if err != nil {
		return
	}

	_, _ = io.WriteString(w, "Upload success: "+newFile.Name())
}

func handleDownload(w http.ResponseWriter, request *http.Request) {
	//文件上传只允许GET方法
	if request.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("Method not allowed"))
		return
	}
	//文件名
	filename := request.FormValue("filename")
	if filename == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Bad request")
		return
	}
	log.Println("filename: " + filename)
	//打开文件
	file, err := os.Open(BaseFilePath + "/" + filename)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Bad request")
		return
	}
	//结束后关闭文件
	defer file.Close()

	//设置响应的header头
	if strings.Contains(filename, "png") {
		w.Header().Add("Content-type", "image/png")
	} else {
		w.Header().Add("Content-type", "application/octet-stream")
		w.Header().Add("Content-Disposition", "attachment;fileName=\""+filename+"\"")
	}
	//将文件写至responseBody
	_, err = io.Copy(w, file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Bad request")
		return
	}
}

func md5Str(byte []byte) string {
	h := md5.New()
	h.Write(byte)
	return hex.EncodeToString(h.Sum(nil))
}

type Config struct {
	APPName string `default:"file-server"`
	DB      struct {
		Name     string
		User     string `default:"root"`
		Password string `required:"true" env:"DBPassword"`
		Port     uint   `default:"3306"`
	}
	Contacts []struct {
		Name  string
		Email string `required:"true"`
	}

	Server struct {
		Port         string `default:"8080"`
		BaseFilePath string `default:"C:/Users/PC/Desktop/tmp/files"`
	}
}

type FileEntity struct {
	Id         int64 `col:"id" json:"id"`
	Hash       string
	Path       string
	Name       string `col:"name" json:"name"`
	Suffix     string
	Status     bool
	CreateTime time.Time
	UpdateTime time.Time
}
