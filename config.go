package storage

import (
	`github.com/storezhang/gox`
)

//  磁盘信息配置
type Config struct {
	// 模拟云盘路径
	Path string `default:"../yunke"`
	// 默认云盘空间大小
	// 单位B
	Size struct {
		// 公共云盘
		Public gox.FileSize `default:"10G"`
		// 私有云盘
		Private gox.FileSize `default:"1G"`
	}

	// 目录层级最大深度
	MaxDeep int `default:"5" yaml:"maxDeep"`
}
