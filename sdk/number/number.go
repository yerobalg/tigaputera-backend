package number

import (
	"golang.org/x/text/language"
    "golang.org/x/text/message"
)

func ConvertToRupiah(number int64) string {
	p := message.NewPrinter(language.Indonesian)
	return p.Sprintf("Rp. %d", number)
}
