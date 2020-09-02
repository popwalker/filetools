package detective

type Detector interface {
	Validate() error
	Detect() error
}

func Detect(d Detector) error {
	return d.Detect()
}
