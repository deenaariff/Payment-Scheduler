package payment_scheduler

import (
	"errors"
	"fmt"
	"math"
	"time"
)

type PaymentScheduler struct{}

const NumInstallments = 3

type TermType string

const TermTypeNet TermType = "net"
const TermTypeInstallments TermType = "installments"

type Currency string

const CurrencyUSD Currency = "USD"

type GetPaymentScheduleParams struct {
	Terms TermType
	// AmountInCents represents total money to be charged in the lowest denomination possible as per Fowler's Money Pattern (https://martinfowler.com/eaaCatalog/money.html)
	AmountInCents int64
	// FeePercentage designates the variable fee rate to be charged per scheduled payment
	FeePercentage int
	// Duration designates the total time length of the payment schedule in days
	Duration int
	// StartDateInMS designates the
	StartDate time.Time
	// Currency represents the currency of the amount being charged in the payment schedule
	Currency Currency
}

func (p GetPaymentScheduleParams) Validate() error {
	if p.Terms == "" {
		return errors.New("must specify a term type")
	}
	if p.AmountInCents <= 0 {
		return errors.New("amount to charge must be greater than 0")
	}
	if p.Terms == TermTypeInstallments && p.AmountInCents < NumInstallments {
		return errors.New(fmt.Sprintf("minimum amount for installments is %v %v", NumInstallments, p.Currency))
	}
	if p.FeePercentage < 0 || p.FeePercentage > 100 {
		return errors.New("fee (in percent) must be an amount between 0 and 100")
	}
	if p.Duration <= 0 {
		return errors.New("duration in days must be greater than 0")
	}
	if p.Currency == "" {
		return errors.New("currency must be specified")
	}
	return nil
}

type ScheduledPayment struct {
	// Date Represents the time at which the payment is charged
	Date time.Time `json:"date"`
	// AmountInCents represents the amount to charged in the scheduled payment in the lowest denomination possible as per Fowler's Money Pattern (https://martinfowler.com/eaaCatalog/money.html)_
	AmountInCents int64 `json:"amountInCents"`
	// Currency represents the currency of the amount being charged in the scheduled payment
	Currency Currency `json:"currency"`
}

func (f PaymentScheduler) GetPaymentSchedule(p GetPaymentScheduleParams) ([]ScheduledPayment, error) {
	err := p.Validate()
	if err != nil {
		return nil, err
	}

	requiresInstallments := p.Terms == TermTypeInstallments

	var remainder int64 // dividing an amount over installments may result in a remainder
	installmentChargeAmount := p.AmountInCents

	if requiresInstallments {
		installmentChargeAmount, remainder = calculateInstallmentAmount(installmentChargeAmount)
	}

	// adjust the installment amount with the fee to be applied
	installmentChargeAmount = applyVariableFee(installmentChargeAmount, p.FeePercentage)
	remainder = applyVariableFee(remainder, p.FeePercentage)

	scheduledPayments := make([]ScheduledPayment, 0)

	if requiresInstallments {
		timeIncrement := p.Duration / (NumInstallments - 1)

		for i := 0; i < NumInstallments-1; i++ {
			newDate := p.StartDate.Add(time.Hour * 24 * time.Duration(i*timeIncrement))

			scheduledPayments = append(scheduledPayments, ScheduledPayment{
				Date:          deferDateToWeekDay(newDate),
				AmountInCents: installmentChargeAmount,
				Currency:      p.Currency,
			})
		}
	}

	endDate := p.StartDate.Add(time.Hour * 24 * time.Duration(p.Duration))

	scheduledPayments = append(scheduledPayments, ScheduledPayment{
		Date:          deferDateToWeekDay(endDate),
		AmountInCents: installmentChargeAmount + remainder,
		Currency:      p.Currency,
	})

	return scheduledPayments, nil
}

func applyVariableFee(amountInCents int64, feeInPercent int) int64 {
	variableRate := float64(feeInPercent) / 100.0
	return int64(math.Ceil(float64(amountInCents) * (1 + variableRate)))
}

func deferDateToWeekDay(date time.Time) time.Time {
	switch date.Weekday() {
	case time.Saturday:
		return date.Add(time.Hour * 24 * time.Duration(2))
	case time.Sunday:
		return date.Add(time.Hour * 24 * time.Duration(1))
	}
	return date
}

func calculateInstallmentAmount(totalAmount int64) (installmentAmount int64, remainder int64) {
	installmentAmount = totalAmount / NumInstallments
	remainder = totalAmount % NumInstallments
	return installmentAmount, remainder
}
