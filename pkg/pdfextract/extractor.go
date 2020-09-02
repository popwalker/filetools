package pdfextract

type Extractor interface {
	Validate() error
	Extract() error
}

func Extract(e Extractor) error {
	return e.Extract()
}
