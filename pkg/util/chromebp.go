package util

import (
	"context"
	"io/ioutil"

	"invtools/utils/errors"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func ChromedpPrintPdf(url string, to string) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, _, err = page.PrintToPDF().
				//WithDisplayHeaderFooter(false).
				//WithLandscape(true).
				//WithPrintBackground(true).
				Do(ctx)
			return err
		}),
	})
	if err != nil {
		return errors.Errorf(err, "chromedp Run failed")
	}

	if err := ioutil.WriteFile(to, buf, 0644); err != nil {
		return errors.Errorf(err, "write to file failed")
	}

	return nil
}
