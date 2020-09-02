package utils

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func CheckFileIsExist(filename string) bool {
	exist := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func CheckDirIsExist(dir string) (exist bool) {
	ret, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return
	}

	if ret != nil && ret.IsDir() {
		exist = true
	}
	return
}

func Mkdir(path string) error {
	if !CheckDirIsExist(path) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

//RmAll 删除当前目录所有文件及其子目录。 相当于os.RemoveAll
func RmAll(path string) error {

	path = strings.TrimSpace(path)

	if path == "" {
		return errors.New(" [RmAll] 目录为空.")
	}

	if path == "/" {
		return fmt.Errorf(" [RmAll] 不能删除根目录。 ")
	}

	//防止路径中存在空格，造成误删除的情况
	if strings.Contains(path, " ") {
		return errors.New(" [RmAll] 路径中包含空格,可能会造成目录误删除. 不作删除处理. path:" + path)
	}

	if path == "." || path == ".." {
		return fmt.Errorf(" [RmAll] 目录信息不对. path: %s", path)
	}

	x, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) { //本身已不存在
			return nil
		}
		return fmt.Errorf(" [RmAll] 得不到path信息. err:%s path:%s ", err, path)
	}

	if !x.IsDir() {
		return fmt.Errorf(" [RmAll] 传入参数不是一个目录. path:%s ", path)
	}

	if IsExcludeDir(path) {
		return fmt.Errorf(" [RmAll] 不能删除特定目录. path:%s", path)
	}

	err = os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf(" [RmAll] 删除path失败.  err:%s path:%s ", err, path)
	}

	return nil
}

//IsExcludeDir 指定目录不能删除
func IsExcludeDir(path string) bool {
	for _, v := range []string{"/", "/var", "/bin", "/dev", "/home", "/lib", "/lib64",
		"/mnt", "/proc", "/opt", "/root", "/sys",
		"/tmp", "/run", "/sbin", "/media", "/etc", "/boot",
		"/srv", "/srv/staticfile", "/srv/staticfile/upload_voucher",
		"/ebs", "/ebs/inv", "/ebs/inv/ticket", "/ebs/inv/ticket/upload",
		"/nfs", "/nfs/staticfile/upload_voucher"} {
		if strings.TrimSpace(strings.TrimSuffix(path, "/")) == v {
			return true
		}
	}

	return false
}

func ComputeMd5String(filePath string) (string, error) {

	var str = ""
	bytes, err := ComputeMd5(filePath)
	if err != nil {
		return str, err
	}

	for _, b := range bytes {
		str += strconv.Itoa(int(b))
	}

	return str, err
}

func ComputeMd5(filePath string) ([]byte, error) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}

func CheckAndMkDir(fdir string) error {
	if !CheckDirIsExist(fdir) {
		err := os.MkdirAll(fdir, 0755)
		if err != nil {
			return fmt.Errorf("[CheckAndMkDir] 目录创建失败. err:%s dir:%s", err, fdir)
		}
	}
	return nil
}
