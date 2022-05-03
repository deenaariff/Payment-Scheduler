package payment_scheduler

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

var (
	testDateJan10, _   = time.Parse("2006-01-02", "2022-01-10")
	testDateJan12, _   = time.Parse("2006-01-02", "2022-01-12")
	testDateFeb9, _    = time.Parse("2006-01-02", "2022-02-09")
	testDateFeb28, _   = time.Parse("2006-01-02", "2022-02-28")
	testDateMarch11, _ = time.Parse("2006-01-02", "2022-03-11")
)

func TestPaymentScheduler_GetPaymentSchedule(t *testing.T) {
	tests := []struct {
		name    string
		params  GetPaymentScheduleParams
		want    []ScheduledPayment
		wantErr error
	}{
		{
			name: "Test invalid amount for installments",
			params: GetPaymentScheduleParams{
				Terms:         TermTypeInstallments,
				AmountInCents: 2,
				FeePercentage: 5,
				Duration:      60,
				StartDate:     testDateJan10,
				Currency:      CurrencyUSD,
			},
			want:    nil,
			wantErr: errors.New("minimum amount for installments is 3 USD"),
		},
		{
			name: "Test Get Schedule Without Increments",
			params: GetPaymentScheduleParams{
				Terms:         TermTypeNet,
				AmountInCents: 3000,
				FeePercentage: 5,
				Duration:      60,
				StartDate:     testDateJan10,
				Currency:      CurrencyUSD,
			},
			want: []ScheduledPayment{
				{
					Date:          testDateMarch11,
					AmountInCents: 3150,
					Currency:      CurrencyUSD,
				},
			},
		},
		{
			name: "Test Get Installments in 60 Day Increments",
			params: GetPaymentScheduleParams{
				Terms:         TermTypeInstallments,
				AmountInCents: 3000,
				FeePercentage: 5,
				Duration:      60,
				StartDate:     testDateJan10,
				Currency:      CurrencyUSD,
			},
			want: []ScheduledPayment{
				{
					Date:          testDateJan10,
					AmountInCents: 1050,
					Currency:      CurrencyUSD,
				},
				{
					Date:          testDateFeb9,
					AmountInCents: 1050,
					Currency:      CurrencyUSD,
				},
				{
					Date:          testDateMarch11,
					AmountInCents: 1050,
					Currency:      CurrencyUSD,
				},
			},
		},
		{
			name: "Test Get Installments in 60 Day Increments With Remainder",
			params: GetPaymentScheduleParams{
				Terms:         TermTypeInstallments,
				AmountInCents: 3001,
				FeePercentage: 5,
				Duration:      60,
				StartDate:     testDateJan10,
				Currency:      CurrencyUSD,
			},
			want: []ScheduledPayment{
				{
					Date:          testDateJan10,
					AmountInCents: 1050,
					Currency:      CurrencyUSD,
				},
				{
					Date:          testDateFeb9,
					AmountInCents: 1050,
					Currency:      CurrencyUSD,
				},
				{
					Date:          testDateMarch11,
					AmountInCents: 1052,
					Currency:      CurrencyUSD,
				},
			},
		},
		{
			name: "Total amount due 45 days from now, adjusted for the weekend",
			params: GetPaymentScheduleParams{
				Terms:         TermTypeNet,
				FeePercentage: 5,
				AmountInCents: 3000,
				Duration:      45,
				StartDate:     testDateJan12,
				Currency:      CurrencyUSD,
			},
			want: []ScheduledPayment{
				{
					Date:          testDateFeb28,
					AmountInCents: 3150,
					Currency:      CurrencyUSD,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := PaymentScheduler{}
			got, err := f.GetPaymentSchedule(tt.params)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPaymentSchedule() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
