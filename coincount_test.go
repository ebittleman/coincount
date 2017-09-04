package coincount

import (
	"math/big"
	"testing"
	"time"
)

func TestCalcCost(t *testing.T) {
	var zero big.Int
	type args struct {
		transactions []InventoryTransaction
		qty          *big.Int
	}
	tests := []struct {
		name     string
		args     args
		wantCost int64
		wantErr  bool
	}{
		{
			name: "does this work",
			args: args{
				transactions: []InventoryTransaction{
					{
						ID:     1,
						QtyIn:  parseEtherFloatToWei(".2"),
						QtyOut: &zero,
						Cost:   350,
					},
					{
						ID:     2,
						QtyIn:  &zero,
						QtyOut: parseEtherFloatToWei(".1"),
						Cost:   350,
					},
					{
						ID:     3,
						QtyIn:  parseEtherFloatToWei(".1"),
						QtyOut: &zero,
						Cost:   300,
					},
					{
						ID:     4,
						QtyIn:  parseEtherFloatToWei("1"),
						QtyOut: &zero,
						Cost:   270,
					},
					{
						ID:     5,
						QtyIn:  &zero,
						QtyOut: parseEtherFloatToWei(".3"),
						Cost:   307,
					},
					{
						ID:     6,
						QtyIn:  parseEtherFloatToWei(".5"),
						QtyOut: &zero,
						Cost:   390,
					},
					{
						ID:     7,
						QtyIn:  parseEtherFloatToWei("6"),
						QtyOut: &zero,
						Cost:   1,
					},
				},
				qty: parseEtherFloatToWei("2.1"),
			},
			wantCost: 210,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCost, err := CalcCost(tt.args.transactions, tt.args.qty)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalcCost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCost != tt.wantCost {
				t.Errorf("CalcCost() = %v, want %v", gotCost, tt.wantCost)
			}
		})
	}
}

func TestAmountCalc(t *testing.T) {
	qty := parseEtherFloatToWei("0.01")
	costOfElectricity := int64(20015)
	p := MiningPayout(time.Unix(123456789, 0), qty, costOfElectricity)
	t.Log(p.Amount)
}
