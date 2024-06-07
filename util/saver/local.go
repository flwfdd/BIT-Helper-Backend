/*
 * @Author: flwfdd
 * @Date: 2024-06-06 16:08:11
 * @LastEditTime: 2024-06-06 22:11:56
 * @Description:
 * _(:з」∠)_
 */
package saver

import (
	"BIT-Helper/util/config"
	"os"
	"path/filepath"
)

// 保存文件到本地 path为子路径
func SaveLocal(path string, content []byte) error {
	// 检查配置开关
	if !config.Config.Saver.Local.Enable {
		return nil
	}

	// 创建路径
	dst := filepath.Join(config.Config.Saver.Local.Path, path)
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	// 写入文件
	os.WriteFile(dst, content, 0666)
	return nil
}
