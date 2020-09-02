package util

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"invtools/utils/errors"
)

func WkHtmlToPDf(reqURL, pdfFile string) error {
	if reqURL == "" || pdfFile == "" {
		return errors.Errorf(nil, "[WkHtmlToPDF] reqURL(%s)或pdfFile(%s)为空.", reqURL, pdfFile)
	}

	URL, err := url.Parse(reqURL)
	if err != nil {
		return errors.Errorf(err, "parse reqURL failed")
	}

	rawQuery := URL.RawQuery
	if strings.Contains(rawQuery, "&") {
		rawQuery = strings.Replace(rawQuery, "&", "\\&", -1)
	}
	reqURL = URL.Scheme + "://" + URL.Host + URL.Path + "?" + rawQuery

	//cmdTpl := `wkhtmltopdf --orientation Portrait --page-size A4 --encoding utf-8 -R 0 -L 0 -T 0 -B 0 --quiet page %s %s`

	cmdTpl := `/usr/local/bin/wkhtmltopdf --orientation Portrait --page-size A4 --encoding utf-8 -R 0 -L 0 -T 0 -B 0 --quiet page %s %s`

	cmdCreate := fmt.Sprintf(cmdTpl, reqURL, pdfFile)
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	err = exec.CommandContext(ctx, "sh", "-c", cmdCreate).Run()
	if err != nil {
		return fmt.Errorf(" err:%s cmd:%s", err, cmdCreate)
	}
	return nil
}
