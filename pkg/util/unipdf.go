package util

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"os"
	"path"
	"strings"

	"invtools/utils/errors"

	rscPdf "github.com/rsc.io/pdf"
	//"github.com/unidoc/unidoc/common/license"
	unisecurity "github.com/unidoc/unipdf/v3/core/security"
	uniextractor "github.com/unidoc/unipdf/v3/extractor"
	unipdf "github.com/unidoc/unipdf/v3/model"
)


type UniPdf struct{}

func NewUniPdf() *UniPdf {
	return &UniPdf{}
}

func (u *UniPdf) ExtractText(inputPath, password string, pages []int) (string, error) {
	// Read input file.
	r, pageCount, _, _, err := readPDF(inputPath, password)
	if err != nil {
		return "", err
	}

	// Extract text.
	if len(pages) == 0 {
		pages = createPageRange(pageCount)
	}

	var text string
	for _, numPage := range pages {
		// Get page.
		page, err := r.GetPage(numPage)
		if err != nil {
			return "", err
		}

		// Extract page text.
		extractor, err := uniextractor.New(page)
		if err != nil {
			return "", err
		}

		pageText, err := extractor.ExtractText()
		if err != nil {
			return "", err
		}

		text += pageText
	}

	return text, nil
}

func (u *UniPdf) ExtractTextWithPages(inputPath, password string, pages []int) ([]string, error) {
	// Read input file.
	r, pageCount, _, _, err := readPDF(inputPath, password)
	if err != nil {
		return nil, err
	}

	// Extract text.
	if len(pages) == 0 {
		pages = createPageRange(pageCount)
	}

	var text []string
	for _, numPage := range pages {
		// Get page.
		page, err := r.GetPage(numPage)
		if err != nil {
			return nil, err
		}

		// Extract page text.
		extractor, err := uniextractor.New(page)
		if err != nil {
			return nil, err
		}

		pageText, err := extractor.ExtractText()
		if err != nil {
			return nil, err
		}

		text = append(text, pageText)
	}

	return text, nil
}

func readPDF(filename, password string) (*unipdf.PdfReader, int, bool, unisecurity.Permissions, error) {
	// Open input file.
	f, err := os.Open(filename)
	if err != nil {
		return nil, 0, false, 0, err
	}
	defer f.Close()

	// Read input file.
	r, err := unipdf.NewPdfReader(f)
	if err != nil {
		return nil, 0, false, 0, err
	}

	// Check if file is encrypted.
	encrypted, err := r.IsEncrypted()
	if err != nil {
		return nil, 0, false, 0, err
	}

	// Decrypt using the specified password, if necessary.
	perms := unisecurity.PermOwner
	if encrypted {
		passwords := []string{password}
		if password != "" {
			passwords = append(passwords, "")
		}

		// Extract use permissions
		_, perms, err = r.CheckAccessRights([]byte(password))
		if err != nil {
			perms = unisecurity.Permissions(0)
		}

		var decrypted bool
		for _, p := range passwords {
			if auth, err := r.Decrypt([]byte(p)); err != nil || !auth {
				continue
			}

			decrypted = true
			break
		}

		if !decrypted {
			return nil, 0, false, 0, errors.New("could not decrypt file with the provided password")
		}
	}

	// Get number of pages.
	pages, err := r.GetNumPages()
	if err != nil {
		return nil, 0, false, 0, err
	}

	return r, pages, encrypted, perms, nil
}

func createPageRange(count int) []int {
	if count <= 0 {
		return []int{}
	}

	var pages []int
	for i := 0; i < count; i++ {
		pages = append(pages, i+1)
	}

	return pages
}

// SplitIntoBytes 拆分成文件
func (u *UniPdf) SplitIntoFiles(filePath string, outputDir string) ([]string, error) {
	var (
		ext      = path.Ext(filePath)
		fileName = strings.TrimSuffix(path.Base(filePath), ext)
	)

	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Errorf(err, "打开文件失败:%s", filePath)
	}
	defer f.Close()

	pdfReader, err := unipdf.NewPdfReaderLazy(f)
	if err != nil {
		return nil, err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, errors.Errorf(err, "check IsEncrypted failed")
	}

	if isEncrypted {
		_, err = pdfReader.Decrypt([]byte(""))
		if err != nil {
			return nil, errors.Errorf(err, "unidoc Decrypt failed")
		}
	}

	unidocNumPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, errors.Errorf(err, "")
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, errors.Errorf(err, "get file stat failed")
	}

	// 使用rsc.io/pdf读取文件
	rscReader, err := rscPdf.NewReader(f, fi.Size())
	if err != nil {
		return nil, err
	}

	// 获取页数
	rscNumPages := rscReader.NumPage()

	// 默认页数以unidoc读取的为准，如果和rscpdf读取的不一致，则以rscpdf读取的页数为准
	numPages := unidocNumPages
	if rscNumPages != unidocNumPages {
		numPages = rscNumPages
	}

	var splittedFilePaths []string
	for i := 1; i <= numPages; i++ {
		pageNum := i

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return nil, errors.Errorf(err, "pdfReader.GetPage 失败")
		}

		// 创建pdfwriter
		pdfWriter := unipdf.NewPdfWriter()
		err = pdfWriter.AddPage(page)
		if err != nil {
			return nil, errors.Errorf(err, "pdfWriter.AddPage failed")
		}

		// 获取拆分后的文件路径
		subFilePath := path.Join(outputDir, fmt.Sprintf("%s_%d%s", fileName, pageNum, ext))

		// 床架文件句柄
		fWrite, err := os.Create(subFilePath)
		if err != nil {
			return nil, errors.Errorf(err, "创建子文件失败")
		}

		// 写入pdf文件
		err = pdfWriter.Write(fWrite)
		if err != nil {
			return nil, errors.Errorf(err, "写入拆分后的pdf文件失败:%s", subFilePath)
		}

		// 关闭文件
		fWrite.Close()
		splittedFilePaths = append(splittedFilePaths, subFilePath)
	}

	return splittedFilePaths, nil
}

// SplitIntoBytes 拆分成Bytes数组
func (u *UniPdf) SplitIntoBytes(filePath string) (map[int][]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Errorf(err, "打开文件失败:%s", filePath)
	}
	defer f.Close()

	pdfReader, err := unipdf.NewPdfReaderLazy(f)
	if err != nil {
		return nil, err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, errors.Errorf(err, "check IsEncrypted failed")
	}

	if isEncrypted {
		_, err = pdfReader.Decrypt([]byte(""))
		if err != nil {
			return nil, errors.Errorf(err, "unidoc Decrypt failed")
		}
	}

	unidocNumPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, errors.Errorf(err, "")
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, errors.Errorf(err, "get file stat failed")
	}

	// 使用rsc.io/pdf读取文件
	rscReader, err := rscPdf.NewReader(f, fi.Size())
	if err != nil {
		return nil, err
	}

	// 获取页数
	rscNumPages := rscReader.NumPage()

	// 默认页数以unidoc读取的为准，如果和rscpdf读取的不一致，则以rscpdf读取的页数为准
	numPages := unidocNumPages
	if rscNumPages != unidocNumPages {
		numPages = rscNumPages
	}

	var splittedFileBytes = make(map[int][]byte)
	for i := 1; i <= numPages; i++ {
		pageNum := i

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return nil, errors.Errorf(err, "pdfReader.GetPage 失败")
		}

		// 创建pdfwriter
		pdfWriter := unipdf.NewPdfWriter()
		err = pdfWriter.AddPage(page)
		if err != nil {
			return nil, errors.Errorf(err, "pdfWriter.AddPage failed")
		}

		var b []byte
		buf := bytes.NewBuffer(b)

		// 写入buffer
		err = pdfWriter.Write(buf)
		if err != nil {
			return nil, errors.Errorf(err, "拆分后的pdf文件写入buffer失败")
		}
		splittedFileBytes[pageNum] = buf.Bytes()
	}

	return splittedFileBytes, nil
}

func (u *UniPdf) ExtractImagesIntoFiles(filePath, outputDir string) ([]string, error) {
	var (
		ext      = path.Ext(filePath)
		fileName = strings.TrimSuffix(path.Base(filePath), ext)
	)

	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Errorf(err, "打开文件失败:%s", filePath)
	}
	defer f.Close()

	pdfReader, err := unipdf.NewPdfReaderLazy(f)
	if err != nil {
		return nil, err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, errors.Errorf(err, "check IsEncrypted failed")
	}

	if isEncrypted {
		_, err = pdfReader.Decrypt([]byte(""))
		if err != nil {
			return nil, errors.Errorf(err, "unidoc Decrypt failed")
		}
	}

	unidocNumPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, errors.Errorf(err, "")
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, errors.Errorf(err, "get file stat failed")
	}

	// 使用rsc.io/pdf读取文件
	rscReader, err := rscPdf.NewReader(f, fi.Size())
	if err != nil {
		return nil, err
	}

	// 获取页数
	rscNumPages := rscReader.NumPage()

	// 默认页数以unidoc读取的为准，如果和rscpdf读取的不一致，则以rscpdf读取的页数为准
	numPages := unidocNumPages
	if rscNumPages != unidocNumPages {
		numPages = rscNumPages
	}

	var imageFiles []string
	for i := 1; i <= numPages; i++ {
		pageNum := i

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return nil, errors.Errorf(err, "pdfReader.GetPage 失败")
		}

		pextract, err := uniextractor.New(page)
		if err != nil {
			return nil, errors.Errorf(err, "new pdf image extractor failed")
		}

		pimages, err := pextract.ExtractPageImages(nil)
		if err != nil {
			return nil, errors.Errorf(err, "pextract.ExtractPageImages failed")
		}
		for idx, img := range pimages.Images {
			fname := fmt.Sprintf("%s_p%d_%d.jpg", fileName, i, idx)

			gimg, err := img.Image.ToGoImage()
			if err != nil {
				return nil, errors.Errorf(err, "img.Image.ToGoImage failed")
			}

			imgFilePath := path.Join(outputDir, fname)
			imgf, err := os.Create(imgFilePath)
			if err != nil {
				return nil, errors.Errorf(err, "创建image文件失败")
			}

			opt := jpeg.Options{Quality: 100}
			err = jpeg.Encode(imgf, gimg, &opt)
			if err != nil {
				return nil, errors.Errorf(err, "jpeg.Encode failed")
			}
			imgf.Close()
			imageFiles = append(imageFiles, imgFilePath)
		}
	}

	return imageFiles, nil
}

func (u *UniPdf) ExtractImagesIntoJpegBytes(filePath string) ([][]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Errorf(err, "打开文件失败:%s", filePath)
	}
	defer f.Close()

	pdfReader, err := unipdf.NewPdfReaderLazy(f)
	if err != nil {
		return nil, err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, errors.Errorf(err, "check IsEncrypted failed")
	}

	if isEncrypted {
		_, err = pdfReader.Decrypt([]byte(""))
		if err != nil {
			return nil, errors.Errorf(err, "unidoc Decrypt failed")
		}
	}

	unidocNumPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, errors.Errorf(err, "")
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, errors.Errorf(err, "get file stat failed")
	}

	// 使用rsc.io/pdf读取文件
	rscReader, err := rscPdf.NewReader(f, fi.Size())
	if err != nil {
		return nil, err
	}

	// 获取页数
	rscNumPages := rscReader.NumPage()

	// 默认页数以unidoc读取的为准，如果和rscpdf读取的不一致，则以rscpdf读取的页数为准
	numPages := unidocNumPages
	if rscNumPages != unidocNumPages {
		numPages = rscNumPages
	}

	var imageBytes [][]byte
	for i := 1; i <= numPages; i++ {
		pageNum := i

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return nil, errors.Errorf(err, "pdfReader.GetPage 失败")
		}

		pextract, err := uniextractor.New(page)
		if err != nil {
			return nil, errors.Errorf(err, "new pdf image extractor failed")
		}

		pimages, err := pextract.ExtractPageImages(nil)
		if err != nil {
			return nil, errors.Errorf(err, "pextract.ExtractPageImages failed")
		}
		for _, img := range pimages.Images {
			gimg, err := img.Image.ToGoImage()
			if err != nil {
				return nil, errors.Errorf(err, "img.Image.ToGoImage failed")
			}
			var b []byte
			buf := bytes.NewBuffer(b)
			opt := jpeg.Options{Quality: 100}
			err = jpeg.Encode(buf, gimg, &opt)
			if err != nil {
				return nil, errors.Errorf(err, "jpeg.Encode failed")
			}

			imageBytes = append(imageBytes, buf.Bytes())
		}
	}

	return imageBytes, nil
}
