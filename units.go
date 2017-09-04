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

func multiplyRoundUp(wei *big.Int, costInCents int64) int64 {
	var remainder big.Int
	centPrecision := big.NewInt(1000)

	amount := big.NewInt(costInCents)
	amount.Mul(amount, centPrecision).
		Mul(amount, wei).
		Div(amount, big.NewInt(weiPerEth))
	remainder.Mod(amount, centPrecision)

	amt := amount.Div(amount, centPrecision).Int64()
	if remainder.Int64() > 1 {
		amt++
	}

	return amt
}

func multiplyTruncate(wei *big.Int, costInCents int64) int64 {
	amount := big.NewInt(costInCents)
	amount.Mul(amount, wei).
		Div(amount, big.NewInt(weiPerEth))

	return amount.Int64()
}

func divideRound(costInCents int64, wei *big.Int) int64 {
	var remainder big.Int
	centPrecision := big.NewInt(1000)

	amount := big.NewInt(costInCents)
	amount.Mul(amount, centPrecision).
		Mul(amount, big.NewInt(weiPerEth)).
		Div(amount, wei)
	remainder.Mod(amount, centPrecision)

	amt := amount.Div(amount, centPrecision).Int64()
	if remainder.Int64() > 1 {
		amt++
	}

	return amt
}

func divideTruncate(costInCents int64, wei *big.Int) int64 {
	amount := big.NewInt(costInCents)
	amount.Mul(amount, big.NewInt(weiPerEth)).
		Div(amount, wei)

	return amount.Int64()
}
