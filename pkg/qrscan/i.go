package qrscan

type Processor interface {
	Extract() error
	Process() error
}
