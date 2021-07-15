package storage

// Storage
type Storage interface {
	// 新建文件
	NewDir(req *NewDirReq) (err error)
	// 新建文件
	NewFile(req *NewFileReq) (err error)
	// 显示目录信息
	ListDir(req *ListDirReq) (rsp *ListDirRsp, err error)
	// 删除文件,目录
	Delete(req *DeleteFileReq) (rsp *DeleteFileRsp, err error)
	// 拷贝文件,目录
	Copy(req *CopyFileReq) (rsp *CopyFileRsp, err error)
	// 重命名文件目录
	Rename(req *RenameFileReq) (rsp *RenameFileRsp, err error)
	// 移动文件
	Move(req *MoveFileReq) (rsp *MoveFileRsp, err error)
}

func New(config Config) (s Storage) {
	s = NewDisk(config)

	return
}
