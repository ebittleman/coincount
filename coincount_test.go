package coincount

import "testing"

func TestCalcCost(t *testing.T) {
	type args struct {
		transactions []InventoryTransaction
		qty          float64
	}
	tests := []struct {
		name     string
		args     args
		wantCost float64
		wantErr  bool
	}{
		{
			name: "does this work",
			args: args{
				transactions: []InventoryTransaction{
					{
						ID:    1,
						QtyIn: .2,
						Cost:  3.5,
					},
					{
						ID:     2,
						QtyOut: .1,
						Cost:   3.5,
					},
					{
						ID:    3,
						QtyIn: .1,
						Cost:  3.0,
					},
					{
						ID:    4,
						QtyIn: 1,
						Cost:  2.7,
					},
					{
						ID:     5,
						QtyOut: .3,
						Cost:   3.0666666666666673,
					},
					{
						ID:    6,
						QtyIn: .5,
						Cost:  3.9,
					},
					{
						ID:    7,
						QtyIn: 6,
						Cost:  1,
					},
				},
				qty: 2.1,
			},
			wantCost: 2.1,
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
