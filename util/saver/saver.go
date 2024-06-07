/*
 * @Author: flwfdd
 * @Date: 2024-06-06 16:07:55
 * @LastEditTime: 2024-06-06 22:26:06
 * @Description:
 * _(:з」∠)_
 */
package saver

import (
	"BIT-Helper/util/config"
	"errors"
	"path/filepath"
	"strings"
)

// 保存文件 返回url
func Save(path string, content []byte) (string, error) {
	err1 := SaveLocal(path, content)
	path = strings.ReplaceAll(path, "\\", "/")
	err2 := SaveCOS(path, content)
	if err1 != nil || err2 != nil {
		return "", errors.New("save failed")
	}
	return GetUrl(path), nil
}

// 通过文件路径获取url
func GetUrl(path string) string {
	return config.Config.Saver.Url + filepath.Join("/", path)
}
