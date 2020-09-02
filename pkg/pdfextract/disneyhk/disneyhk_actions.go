package disneyhk

type SequentialNo = string
type ReservationCode = string

const (
	FieldDescSequentialNo    = "SequentialNo"
	FieldDescReservationCode = "ReservationCode"
)

// VoucherPageInfo 一页凭证所需的信息
type VoucherPageInfo struct {
	SequentialNo    SequentialNo
	ReservationCode ReservationCode
}

type Actions []Action
type Action func(input string) (string, error)

func SequentialNoAction(input string) (SequentialNo, error) {
	return "", nil
}
func ReservationCodeAction(input string) (ReservationCode, error) {
	return "", nil
}

var defaultActions = Actions{SequentialNoAction, ReservationCodeAction}
