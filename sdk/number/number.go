package number

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"math"
)

func ConvertToRupiah(number int64) string {
	p := message.NewPrinter(language.Indonesian)
	return p.Sprintf("Rp. %d", number)
}

func GetPercentage(number int64, total int64) float64 {
	if total == 0 {
		return 0
	}

	return roundFloat(float64(number)/float64(total)*100, 2)
}

func roundFloat(number float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return math.Round(number*output) / output
}
