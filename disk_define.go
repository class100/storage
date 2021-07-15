package storage

import (
	`github.com/storezhang/gox`
)

const (
	// 公共
	StorageTypePublic  storageType = "public"
	// 私人
	StorageTypePrivate   = "private"
)

type (
	storageType string

	//  基础请求
	BaseDisk struct {
		// Type 根据不同的编号获取不同的地址
		// private 个人云盘
		// public 公共云盘
		Type storageType `json:"type" validate:"required,oneof= public private"`
		// Id 编号
		// 当Type为private时 为用户编号
		// 当Type为public 时 为机构编号
		Id int64 `json:"id,string" validate:"required"`
	}

	// 创建新的目录,文件
	NewDirReq struct {
		BaseDisk

		// Path 路径
		Path string `json:"path" validate:"required,startswith=/,filename"`
	}

	// 创建新的目录,文件
	NewFileReq struct {
		BaseDisk

		// Path 路径
		Path string `json:"path" validate:"required,startswith=/,filename"`
	}

	// 获取文件列表请求
	ListDirReq struct {
		BaseDisk

		// Path 父路径
		Path string `json:"path" validate:"required,startswith=/,filename"`
	}

	// 获取文件列表响应
	ListDirRsp struct {
		// Infos 文件信息
		Infos []*dirFileInfo `json:"infos"`
	}

	// 删除文件列表请求
	DeleteFileReq struct {
		BaseDisk

		// path 文件路径
		// 文件全路径,从根路径开始比如/a/b.txt
		Path string `json:"path" validate:"required,startswith=/"`
	}

	DeleteFileRsp struct {
		// 受影响的文件
		EffectedFiles []*effectedFile `json:"effectedFiles"`
		// 受影响的文件编号
		FileIds   []string `json:"fileIds"`
	}

	// 拷贝文件请求
	CopyFileReq struct {
		BaseDisk

		// Src 文件路径
		// 使用全量路径 从根路径开始 比如/a/b.txt
		Src string `json:"src" validate:"required,startswith=/"`
		// Dst 文件目的路径
		// 从根路径开始 比如/a/b.txt
		Dst string `json:"dst" validate:"required,startswith=/"`
	}

	// 拷贝文件数据
	CopyFileRsp struct {
		// 受影响的文件编号
		FileIds   []string `json:"fileIds"`
	}

	//  重命名文件请求
	RenameFileReq struct {
		BaseDisk

		// 修改类型
		// 1 文件夹
		// 2 文件
		FileType gox.FileType `json:"fileType" validate:"required,oneof=1 2"`
		// Path 文件夹/文件的父路径
		// 例如：/
		Path string `json:"path" validate:"required,startswith=/"`
		//  要更改的名字
		// 文件夹	例如：b
		// 文件  	例如: b.txt
		Src string `json:"src" validate:"required"`
		// 目的名字
		// 文件夹	例如：c
		// 文件  	例如: c.txt
		Dst string `json:"dst" validate:"required,filename"`
		// FileId 文件唯一Id
		FileId string `json:"fileId"`
		// Desc 描述
		Desc string `json:"desc"`
	}

	RenameFileRsp struct {
		// 受影响的文件
		EffectedFiles []*effectedFile `json:"effectedFiles"`
	}

	// 移动文件请求
	MoveFileReq struct {
		BaseDisk

		// 文件路径
		// 使用全量路径 从根路径开始 比如/a/b.txt
		Src string `json:"src" validate:"required,startswith=/"`
		// 文件目的路径
		// 从根路径开始 比如/a/b.txt
		Dst string `json:"dst" validate:"required,startswith=/"`
	}

	MoveFileRsp struct {
		// 受影响的文件
		EffectedFiles []*effectedFile `json:"effectedFiles"`
	}

	//// 存储信息請求
	//GetInfoReq struct {
	//	BaseDisk
	//}
	//
	//// 存储信息响应
	//GetInfoRsp struct {
	//	// 默认总大小 单位B
	//	DefaultSize int64 `json:"defaultSize"`
	//	// 使用大小 单位B
	//	UseSize int64 `json:"useSize"`
	//}

	// 文件信息
	dirFileInfo struct {
		// 文件类型
		// 目录：1
		// 文件：2
		Type gox.FileType `json:"type"`
		// 文件路径
		Path string `json:"path"`
		// 更新时间
		UpdateTime gox.Timestamp `json:"updateTime"`
	}

	// 影响的目录
	effectedFile struct {
		// 改变前
		Src string
		// 改变后
		Dst string
	}
)