package saver

import (
	"BIT-Helper/util/config"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/tencentyun/cos-go-sdk-v5"
)

var client *cos.Client
var once sync.Once

func InitCOS() {
	once.Do(func() {
		u, _ := url.Parse(fmt.Sprintf("https://%v.cos.%v.myqcloud.com", config.Config.Saver.Cos.Bucket, config.Config.Saver.Cos.Region))
		b := &cos.BaseURL{BucketURL: u}

		client = cos.NewClient(b, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  config.Config.Saver.Cos.SecretId,
				SecretKey: config.Config.Saver.Cos.SecretKey,
			},
		})
	})
}

// 保存文件到腾讯云COS
func SaveCOS(path string, data []byte) error {
	if !config.Config.Saver.Cos.Enable {
		return nil
	}
	InitCOS()
	_, err := client.Object.Put(context.Background(), config.Config.Saver.Cos.Path+path, bytes.NewReader(data), nil)
	if err != nil {
		return err
	}
	return nil
}
