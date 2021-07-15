package storage

import (
	`fmt`
	`io/ioutil`
	`os`
	`path/filepath`
	`sort`
	`strings`

	`github.com/storezhang/gox`
)

type Disk struct {
	config Config
}

func NewDisk(config Config) *Disk {
	return &Disk{config: config}
}

func (d *Disk) NewDir(req *NewDirReq) (err error) {
	req.Path = strings.TrimRight(req.Path, "/")
	if strings.Count(req.Path, "/") > d.config.MaxDeep {
		err = ErrDirMaxDeep

		return
	}

	path := d.getPath(req.BaseDisk, req.Path)

	if err = d.checkCreateFile(path); nil != err {
		return
	}

	if err = gox.CreateDir(path); nil != err {
		return
	}

	return
}

func (d *Disk) NewFile(req *NewFileReq) (err error) {
	path := d.getPath(req.BaseDisk, req.Path)
	if err = gox.CreateFile(path); nil != err {
		return
	}

	return
}

// 显示目录信息
func (d *Disk) ListDir(req *ListDirReq) (rsp *ListDirRsp, err error) {
	var (
		files  []*dirFileInfo
		prefix = req.Path
		path   = d.getPath(req.BaseDisk, req.Path)
	)

	if prefix == "/" {
		if err = gox.DirNotExistCreate(path); nil != err {
			return
		}
	}

	var fs []os.FileInfo
	fs, err = ioutil.ReadDir(path)
	if nil != err {
		return nil, ErrListDirParamIsNotDir
	}

	for _, fi := range fs {
		p := filepath.ToSlash(filepath.Join(prefix, fi.Name()))
		if fi.IsDir() {
			files = append(files, &dirFileInfo{
				Type:       gox.FileTypeDir,
				Path:       p,
				UpdateTime: gox.ParseTimestamp(fi.ModTime()),
			})
		} else {
			files = append(files, &dirFileInfo{
				Type:       gox.FileTypeFile,
				Path:       p,
				UpdateTime: gox.ParseTimestamp(fi.ModTime()),
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].Type == files[j].Type {
			return files[i].UpdateTime.Time().Sub(files[j].UpdateTime.Time()) > 0
		}

		return gox.FileTypeDir == files[i].Type
	})

	rsp = &ListDirRsp{Infos: files}

	return
}

// 删除文件,目录
func (d *Disk) Delete(req *DeleteFileReq) (rsp *DeleteFileRsp, err error) {
	var (
		deletedFiles []string
	)

	rsp = new(DeleteFileRsp)

	path := d.getPath(req.BaseDisk, req.Path)

	if deletedFiles, err = gox.GetNeedDeleteFiles(path); nil != err {
		return
	}

	if StorageTypePublic == req.Type {
		rsp.EffectedFiles = d.deleteEffectedFileDir(req)
	}

	if err = gox.DeleteFile(path); nil != err {
		return
	}

	for _, f := range deletedFiles {
		var fileId string

		_, fn := filepath.Split(f)
		if fileId, _, err = d.splitFilename(fn); nil != err {
			continue
		}
		rsp.FileIds = append(rsp.FileIds, fileId)
	}

	return
}

// 拷贝文件,目录
func (d *Disk) Copy(req *CopyFileReq) (rsp *CopyFileRsp, err error) {
	var copyFiles []string

	rsp = new(CopyFileRsp)
	if err = d.checkCanCopyMove(req.BaseDisk, req.Src, req.Dst); nil != err {
		return
	}

	dstPath := d.getPath(req.BaseDisk, req.Dst)
	srcPath := d.getPath(req.BaseDisk, req.Src)

	if copyFiles, err = gox.CopyFile(srcPath, dstPath); nil != err {
		return
	}
	for _, f := range copyFiles {
		var fileId string

		_, fn := filepath.Split(f)
		if fileId, _, err = d.splitFilename(fn); nil != err {
			return
		}
		rsp.FileIds = append(rsp.FileIds, fileId)
	}

	return
}

// 重命名文件目录
func (d *Disk) Rename(req *RenameFileReq) (rsp *RenameFileRsp, err error) {
	var (
		path    string
		srcPath string
		dstPath string
	)

	rsp = new(RenameFileRsp)
	path = d.getPath(req.BaseDisk, req.Path)

	if gox.FileTypeDir == req.FileType {
		srcPath = filepath.ToSlash(filepath.Join(path, req.Src))
		dstPath = filepath.ToSlash(filepath.Join(path, req.Dst))
	} else if gox.FileTypeFile == req.FileType {
		srcName := fmt.Sprintf("%v_%v_%v", req.FileId, req.Desc, req.Src)
		desName := fmt.Sprintf("%v_%v_%v", req.FileId, req.Desc, req.Dst)
		srcPath = filepath.ToSlash(filepath.Join(path, srcName))
		dstPath = filepath.ToSlash(filepath.Join(path, desName))
	}

	if gox.IsFileExist(dstPath) {
		err = &gox.CodeError{
			ErrorCode: ErrFilenameSameName.ErrorCode,
			Message:   fmt.Sprintf(ErrFilenameSameName.Message, req.Dst),
			Data:      ErrFilenameSameName.Data,
		}

		return
	}

	rsp.EffectedFiles = d.renameEffectedFileDir(req, srcPath, dstPath)

	if err = gox.RenameFile(srcPath, dstPath); nil != err {
		return
	}

	return
}

// 移动文件
func (d *Disk) Move(req *MoveFileReq) (rsp *MoveFileRsp, err error) {
	if err = d.checkCanCopyMove(req.BaseDisk, req.Src, req.Dst); nil != err {
		return
	}
	rsp = new(MoveFileRsp)

	dstPath := d.getPath(req.BaseDisk, req.Dst)
	srcPath := d.getPath(req.BaseDisk, req.Src)
	rsp.EffectedFiles = d.moveEffectedFileDir(req.BaseDisk, srcPath, dstPath)

	if err = gox.MoveFile(srcPath, dstPath); nil != err {
		return
	}

	return
}

func (d *Disk) deleteEffectedFileDir(req *DeleteFileReq) (effectedFiles []*effectedFile) {
	var (
		srcPath      = d.getPath(req.BaseDisk, req.Path)
		defaultPath  = d.getDefaultPath(req.BaseDisk)
		err          error
		changedFiles []string
		src          = strings.TrimPrefix(srcPath, defaultPath)
	)

	if isDir, _ := gox.IsDir(srcPath); isDir {
		if changedFiles, err = gox.GetAllFilesBy(srcPath, gox.FileTypeAny); nil != err {
			return
		}
	} else {
		changedFiles = append(changedFiles, src)
	}
	effectedFiles = make([]*effectedFile, 0, len(changedFiles))
	effectedFiles = append(effectedFiles, &effectedFile{
		Src: src,
		Dst: "",
	})

	for _, cd := range changedFiles {
		cd = filepath.ToSlash(cd)
		srcName := strings.TrimPrefix(cd, defaultPath)
		effectedFiles = append(effectedFiles, &effectedFile{
			Src: srcName,
			Dst: "",
		})
	}

	return
}

func (d *Disk) renameEffectedFileDir(req *RenameFileReq, srcPath, dstPath string, ) (effectedFiles []*effectedFile) {
	var (
		defaultPath  = d.getDefaultPath(req.BaseDisk)
		err          error
		changedFiles []string
	)

	if gox.FileTypeDir == req.FileType {
		if changedFiles, err = gox.GetAllFilesBy(srcPath, gox.FileTypeAny); nil != err {
			return
		}
		effectedFiles = make([]*effectedFile, 0, len(changedFiles))
		for _, cd := range changedFiles {
			cd = filepath.ToSlash(cd)
			srcName := strings.TrimPrefix(cd, defaultPath)
			dstName := strings.TrimPrefix(strings.Replace(cd, srcPath, dstPath, 1), defaultPath)
			effectedFiles = append(effectedFiles, &effectedFile{
				Src: srcName,
				Dst: dstName,
			})
		}
		srcName := strings.TrimPrefix(srcPath, defaultPath)
		dstName := strings.TrimPrefix(dstPath, defaultPath)
		effectedFiles = append(effectedFiles, &effectedFile{
			Src: srcName,
			Dst: dstName,
		})
	} else {
		effectedFiles = append(effectedFiles, &effectedFile{
			Src: strings.TrimPrefix(srcPath, defaultPath),
			Dst: strings.TrimPrefix(dstPath, defaultPath),
		})
	}

	return
}

func (d *Disk) moveEffectedFileDir(disk BaseDisk, srcPath, dstPath string, ) (effectedFiles []*effectedFile) {
	var (
		defaultDiskPath = d.getDefaultPath(disk)
		err             error
		changedFiles    []string
		srcName         string
	)

	if isDir, _ := gox.IsDir(srcPath); isDir {
		if changedFiles, err = gox.GetAllFilesBy(srcPath, gox.FileTypeAny); nil != err {
			return
		}
		src := strings.TrimPrefix(srcPath, defaultDiskPath)
		idx := strings.LastIndex(src, "/")
		if -1 != idx {
			srcName = src[idx+1:]
		}
		replaceDst := filepath.ToSlash(filepath.Join(dstPath, srcName))
		dst := strings.TrimPrefix(replaceDst, defaultDiskPath)

		effectedFiles = make([]*effectedFile, 0, len(changedFiles))
		effectedFiles = append(effectedFiles, &effectedFile{
			Src: src,
			Dst: dst,
		})

		for _, cd := range changedFiles {
			cd = filepath.ToSlash(cd)
			srcName := strings.TrimPrefix(cd, defaultDiskPath)
			dstName := strings.TrimPrefix(strings.Replace(cd, srcPath, replaceDst, 1), defaultDiskPath)
			effectedFiles = append(effectedFiles, &effectedFile{
				Src: srcName,
				Dst: dstName,
			})
		}
	} else {
		_, fileName := filepath.Split(srcPath)
		effectedFiles = append(effectedFiles, &effectedFile{
			Src: strings.TrimPrefix(srcPath, defaultDiskPath),
			Dst: strings.TrimPrefix(filepath.ToSlash(filepath.Join(dstPath, fileName)), defaultDiskPath),
		})
	}

	return
}

func (d *Disk) getPath(disk BaseDisk, filePath string) (path string) {
	return fmt.Sprintf("%v/%v/%d%v", d.config.Path, disk.Type, disk.Id, filePath)
}

func (d *Disk) getDefaultPath(disk BaseDisk) (path string) {
	return fmt.Sprintf("%v/%v/%d", d.config.Path, disk.Type, disk.Id)
}

func (d *Disk) splitFilename(path string) (id string, fileName string, err error) {
	_, file := filepath.Split(path)
	items := strings.SplitN(file, "_", 3)
	if 3 != len(items) {
		err = ErrFileNameFormat

		return
	}
	id = items[0]
	fileName = items[2]

	return
}

func (d *Disk) getSrcDstName(src, dst string) (srcName, dstName string) {
	idx := strings.LastIndex(src, "/")
	if -1 != idx {
		srcName = src[idx+1:]
	}

	dstName = filepath.ToSlash(filepath.Join(dst, srcName))

	return
}

func (d *Disk) checkCreateFile(path string) (err error) {
	var newName string

	idx := strings.LastIndex(path, "/")
	if -1 != idx {
		newName = path[idx+1:]
	}

	if gox.IsFileExist(path) {
		err = &gox.CodeError{
			ErrorCode: ErrFilenameSameName.ErrorCode,
			Message:   fmt.Sprintf(ErrFilenameSameName.Message, newName),
			Data:      ErrFilenameSameName.Data,
		}

		return
	}

	return
}

func (d *Disk) checkCanCopyMove(disk BaseDisk, src string, dst string) (err error) {
	dstFatherDeep := gox.GetDirFatherDeep(dst)

	if strings.HasPrefix(dst, src) {
		err = ErrMoveCopyDstSubDirOfSrc

		return
	}

	srcPath := d.getPath(disk, src)
	srcName, dstPathName := d.getSrcDstName(src, dst)
	srcDstPath := d.getPath(disk, dstPathName)
	if gox.IsFileExist(srcDstPath) {
		err = &gox.CodeError{
			ErrorCode: ErrFilenameSameName.ErrorCode,
			Message:   fmt.Sprintf(ErrFilenameSameName.Message, srcName),
			Data:      ErrFilenameSameName.Data,
		}

		return
	}

	srcSonDeep := gox.GetDirSonDeep(srcPath)

	if (dstFatherDeep + srcSonDeep) > d.config.MaxDeep {
		err = ErrDirMaxDeep

		return
	}

	return
}
