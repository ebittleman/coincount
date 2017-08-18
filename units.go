package coincount

import (
	"math"
	"math/big"
	"strings"
)

var weiPerEth = int64(math.Pow(10, 18))

func parseEtherFloatToWei(amount string) *big.Int {
	parts := strings.Split(amount, ".")
	weiStr := "0"
	if len(parts) == 2 {
		weiStr = padRight(parts[1], "0", 18)
	}

	var wei big.Int
	wei.SetString(weiStr, 10)

	if parts[0] == "" {
		parts[0] = "0"
	}
	var ether big.Int
	ether.SetString(parts[0], 10)

	total := big.NewInt(weiPerEth)
	total.Mul(total, &ether)
	total.Add(total, &wei)

	return total
}

func padRight(str, pad string, length int) string {
	for {
		str += pad
		if len(str) > length {
			return str[0:length]
		}
	}
}
