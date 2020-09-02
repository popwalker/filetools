### filetools

filetools is a command line tool which build for increasing daily work productivity powered by [cobra](https://github.com/spf13/cobra).

The idea to build this tool came from my daily work, for details please see my article [Process PDF with Golang](https://www.reddit.com/r/golang/comments/eihcts/process_pdf_with_golang/)

### Support:
- generate pdf from URL
- detect information from pdf
- extract infrmation from pdf
- repair broken pdf
- scan qrcode/barcode

### Compile
```
go build -o filetools
```

> Because of some CGO issue, It doesn't support windows now, but work fine with macOS and Linux.


### Usage:
```
$ filetools

filetools is a command line tool which build for increasing team productivity.
Build by @cxf.

Usage:
  filetools [command]

Available Commands:
  help        Help about any command
  linktopdf   Print PDF from link(url).
  pdfdetect   Pdfdetect is a tool to detect/analyze pdf.
  pdfextract  Extract information from pdf vouchers.
  pdfrepair   Repair PDF files
  pdfsplit    Split PDF files.
  qrcodescan  Scan QRCode/Barcode

Flags:
  -h, --help      help for filetools
      --version   version for filetools

Use "filetools [command] --help" for more information about a command.
```


Reference:
- [wkhtmltopdf](https://wkhtmltopdf.org/)
- [xpdf](https://www.xpdfreader.com/)
- [unipdf](https://unidoc.io/)
- [pdflib/tet](https://www.pdflib.com/products/tet/)
- [mupdf](https://www.mupdf.com/index.html)
- [pdfcpu](https://pdfcpu.io/)
- [chromedp](https://github.com/chromedp/chromedp)
- [gift](github.com/disintegration/gift)
