package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/jinzhu/configor"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var (
	BaseFilePath string
	ServerPort   string
	Host         string
	db           *sql.DB
	conf         Config
	imageExt     string
	minioClient  *minio.Client
)

func main() {

	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/download", handleDownload)
	http.HandleFunc("/s3/upload", handleUploadMinIO)
	http.HandleFunc("/s3/getPreSignedPutUrl", getPreSignedPutUrl)

	err := http.ListenAndServe(ServerPort, nil)
	if err != nil {
		log.Fatal("Server run fail")
	}
}

// 1.配置文件初始化
// 2.数据库初始化
func init() {
	// 配置文件
	conf = Config{}
	err := configor.Load(&conf, "config.yml")
	if err != nil {
		panic(err)
	}
	BaseFilePath = conf.Server.BaseFilePath
	ServerPort = ":" + conf.Server.Port
	Host = conf.Server.Host
	imageExt = conf.ImageExt

	fmt.Printf("%v \n", conf)

	if conf.Minio.Enable {
		minioClient, err = minio.New(conf.Minio.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(conf.Minio.Id, conf.Minio.Secret, ""),
			Secure: conf.Minio.Secure,
		})
	}

	// 数据库
	if conf.DB.Enable {
		//connStr := "postgres://postgres:postgres@192.168.122.11/file-server?sslmode=disable"
		connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", conf.DB.User, conf.DB.Password, conf.DB.Host, conf.DB.Port, conf.DB.Name)
		db, err = sql.Open("postgres", connStr)
	}

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

	fileEntity := FileEntity{
		Hash: md5str,
		Path: newFile.Name(),
	}
	insertFile(fileEntity)
	m := make(map[string]string)
	m["hash"] = md5str
	m["url"] = "http://" + Host + ServerPort + "/download?filename=" + md5str + fileHeader.Filename

	_, _ = io.WriteString(w, success(m))
}

func handleDownload(w http.ResponseWriter, request *http.Request) {
	//文件下载只允许GET方法
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
	fileType := path.Ext(filename)
	if strings.Contains(imageExt, fileType) {
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

func handleUploadMinIO(w http.ResponseWriter, request *http.Request) {

	fmt.Println("path", request.URL.Path)
	fmt.Println("scheme", request.URL.Scheme)
	fmt.Println(request.Form["url_long"])

	//文件上传只允许PUT方法
	if request.Method != http.MethodPut {
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
	var contentType string
	fileName := fileHeader.Filename
	fileType := path.Ext(fileName)
	if strings.Contains(imageExt, fileType) {
		contentType = "image/jpg"
	} else {
		contentType = "application/octet-stream"
	}
	_, err = minioClient.PutObject(context.Background(), "image", fileName, file, fileHeader.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		_, _ = io.WriteString(w, err.Error())
		return
	}
	//defer 结束时关闭文件
	defer file.Close()
	log.Println("filename: " + fileName)

	m := make(map[string]string)
	m["url"] = conf.Minio.Endpoint + "/image/" + fileName

	_, _ = io.WriteString(w, success(m))
}

// 预签名上传文件url
func getPreSignedPutUrl(w http.ResponseWriter, request *http.Request) {
	client, err := minioClient.PresignedPutObject(context.Background(), "image", uuid.New().String(), 1*time.Minute)
	if err != nil {
		_, _ = io.WriteString(w, "Read file error")
		return
	}
	_, _ = io.WriteString(w, success(client.String()))
}
