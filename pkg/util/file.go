package util

import (
	"io/ioutil"
	"path"
	"strings"

	"invtools/utils"

	"invtools/utils/errors"
)

var (
	// uselessDirs 需要过滤掉的路径
	uselessDirs = []string{"__MACOSX"}
	// uselessFiles 需要过滤掉的文件
	uselessFiles = []string{".DS_Store", "Thumbs.db"}
)

// IsUselessDir return if dirname is useless
func IsUselessDir(dirname string) bool {
	for _, v := range uselessDirs {
		if dirname == v {
			return true
		}
	}
	return false
}

// return if filename is useless
func IsUselessFile(filename string) bool {
	for _, v := range uselessFiles {
		if filename == v {
			return true
		}
	}
	return false
}

// ReadDirFiles get all files under dir with expected extension
func ReadDirFiles(dir, ext string) ([]string, error) {
	if ok := utils.CheckDirIsExist(dir); !ok {
		return nil, errors.Errorf(nil, "directory not exists,directory:%s", dir)
	}

	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Errorf(err, "read directory failed,directory:%s", dir)
	}

	var fs []string
	for _, fi := range fis {
		// ignore directory
		if fi.IsDir() {
			continue
		}

		// ignore unexpected file extension
		if strings.ToLower(path.Ext(fi.Name())) != ext {
			continue
		}

		fs = append(fs, path.Join(dir, fi.Name()))
	}

	return fs, nil
}

func ReadDirFilesV2(dir string) ([]string, error) {
	if ok := utils.CheckDirIsExist(dir); !ok {
		return nil, errors.Errorf(nil, "directory not exists,directory:%s", dir)
	}

	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Errorf(err, "read directory failed,directory:%s", dir)
	}

	var fs []string
	for _, fi := range fis {
		// ignore directory
		if fi.IsDir() {
			continue
		}

		fs = append(fs, path.Join(dir, fi.Name()))
	}

	return fs, nil
}

// ReadDirFilesV3 递归读取文件目录，获取所有文件,指定扩展名
func ReadDirFilesV3(dirname string, ext ...string) ([]string, error) {
	if dirname == "" {
		return nil, errors.Errorf(nil, "param directory cannot be empty")
	}

	fis, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, errors.Errorf(err, "readDir:%s failed", dirname)
	}

	var uniqueFiles = make(map[string]struct{})
	var result []string
	for _, fi := range fis {
		// 过滤掉一些无用的路径
		if fi.IsDir() && IsUselessDir(fi.Name()) {
			continue
		}
		if fi.IsDir() {
			subDirname := path.Join(dirname, fi.Name())
			fps, err := ReadDirFilesV3(subDirname, ext...)
			if err != nil {
				return nil, errors.Errorf(err, "read dir:%s failed", subDirname)
			}
			result = append(result, fps...)
			continue
		}

		// 只匹配指定扩展名的文件
		for _, e := range ext {
			if !strings.HasSuffix(strings.ToLower(fi.Name()), e) {
				continue
			}
		}

		fp := path.Join(dirname, fi.Name())
		if _, ok := uniqueFiles[fp]; !ok {
			uniqueFiles[fp] = struct{}{}
		}
	}

	// macOS创建的tar.gz文件解压后会存在以 "._"开头的隐藏文件，这里做一次过滤
	// eg: uniqueFiles = map[string]struct{}{"file_a.pdf": {}, "._file_a.pdf": {}}
	for fp, _ := range uniqueFiles {
		fName := path.Base(fp)
		fDir := path.Dir(fp)
		if strings.HasPrefix(fName, "._") {
			cleanFName := strings.TrimPrefix(fName, "._")
			cleanFPath := path.Join(fDir, cleanFName)
			if _, ok := uniqueFiles[cleanFPath]; ok {
				continue
			}
		}

		// 过滤掉无用的文件
		if IsUselessFile(fName) {
			continue
		}
		result = append(result, fp)
	}

	return result, nil
}

func GetPureFileName(filePath string) string {
	return strings.TrimSuffix(path.Base(filePath), path.Ext(filePath))
}
