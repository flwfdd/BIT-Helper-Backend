/*
 * @Author: flwfdd
 * @Date: 2024-06-06 15:59:19
 * @LastEditTime: 2024-06-07 12:27:10
 * @Description:
 * _(:з」∠)_
 */
package config

import (
	"fmt"
	"os"

	"github.com/jinzhu/configor"
)

var Config = struct {
	Port        string
	Key         string
	LoginExpire int64 `yaml:"login_expire"`
	Dsn         string
	Saver       struct {
		MaxSize        int64 `yaml:"max_size"`
		Url            string
		ImageUrlSuffix string `yaml:"image_url_suffix"`
		Local          struct {
			Enable bool
			Path   string
		}
		Cos struct {
			Enable    bool
			SecretId  string `yaml:"secret_id"`
			SecretKey string `yaml:"secret_key"`
			Bucket    string
			Region    string
			Path      string
		}
	}
	DefaultAvatar string `yaml:"default_avatar"`
	PageSize      int    `yaml:"page_size"`
	ReleaseMode   bool   `yaml:"release_mode"`
}{}

func Init() {
	path := "config.yml"
	_, err := os.Stat(path)
	if err != nil {
		fmt.Println("config.yml not found, please copy config_example.yml to config.yml and edit it")
		os.Exit(1)
	}
	configor.Load(&Config, path)
	// fmt.Printf("config: %#v", Config)
}
