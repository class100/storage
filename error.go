package storage

import (
	`github.com/storezhang/gox`
)

var (
	// 文件相关
	ErrListDirParamIsNotDir   = &gox.CodeError{ErrorCode: 9102, Message: "获取文件信息参数不是目录"}
	ErrFileNameFormat         = &gox.CodeError{ErrorCode: 9103, Message: "错误文件名格式"}
	ErrMoveCopyDstSubDirOfSrc = &gox.CodeError{ErrorCode: 9105, Message: "目的路径不能是源路径或是源路径的子目录"}
	ErrDirMaxDeep             = &gox.CodeError{ErrorCode: 9106, Message: "目录深度最大为5层"}
	ErrFilenameSameName       = &gox.CodeError{ErrorCode: 9108, Message: "%v名字相同操作失败"}
	ErrFileChunkNotFound      = &gox.CodeError{ErrorCode: 9109, Message: "分块上传文件不存在"}
)