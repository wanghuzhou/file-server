package main

import "time"

type Config struct {
	APPName string `default:"file-server"`
	DB      struct {
		Name     string
		User     string `default:"root"`
		Password string `required:"true" env:"DBPassword"`
		Host     string `default:"localhost"`
		Port     uint   `default:"5432"`
		Use      bool   `default:"true"`
	}
	Contacts []struct {
		Name  string
		Email string `required:"true"`
	}

	Server struct {
		Port         string `default:"8080"`
		BaseFilePath string `default:"./tmp" yaml:"baseFilePath"`
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

type Object struct{}

type UploadVO struct {
	Object
	Hash string
	Url  string
}

type ResultDTO struct {
	Code int32 `json:"code"`
	Msg  string
	Data interface{}
}
