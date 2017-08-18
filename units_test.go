package coincount

import (
	"math/big"
	"reflect"
	"testing"
)

func Test_parseEtherFloatToWei(t *testing.T) {
	type args struct {
		amount string
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		{
			name: "Test1",
			args: args{
				amount: "0.000000000000000001",
			},
			want: big.NewInt(1),
		},
		{
			name: "Test1",
			args: args{
				amount: "0.999999999999999999",
			},
			want: big.NewInt(999999999999999999),
		},
		{
			name: "Test2",
			args: args{
				amount: ".999999999999999999",
			},
			want: big.NewInt(999999999999999999),
		},
		{
			name: "Test3",
			args: args{
				amount: "999999999999999999.999999999999999999",
			},
			want: func() *big.Int {
				ether := big.NewInt(999999999999999999)

				wei := big.NewInt(999999999999999999)
				return ether.Mul(ether, big.NewInt(weiPerEth)).
					Add(ether, wei)
			}(),
		},
		{
			name: "Test4",
			args: args{
				amount: "999999999999999999.999999999999999999",
			},
			want: func() *big.Int {
				var ether big.Int
				ether.SetString("999999999999999999999999999999999999", 10)

				return &ether
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseEtherFloatToWei(tt.args.amount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEtherFloatToWei() = %v, want %v", got, tt.want)
			} else {
				t.Log(got.String())
			}
		})
	}
}
