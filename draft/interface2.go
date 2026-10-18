package main

import "fmt"

type Wallet struct {
	Cash int
}

func (w *Wallet) Pay(amount int) error {
	if w.Cash < amount {
		return fmt.Errorf("not enough cash")
	}
	w.Cash -= amount
	return nil
}

type Card struct {
	Balance    int
	ValidUntil string
	Cardholder string
	CVV        string
	Number     string
}

func (c *Card) Pay(amount int) error {
	if c.Balance < amount {
		return fmt.Errorf("Not enough money on balance")
	}
	c.Balance -= amount
	return nil
}

type ApplePay struct {
	Money   int
	AppleID string
}

func (a *ApplePay) Pay(amount int) error {
	if a.Money < amount {
		return fmt.Errorf("Not enough money on account")
	}
	a.Money -= amount
	return nil
}

type Payer interface {
	Pay(amount int) error
}

func Buy(p Payer) {
	switch p.(type) {
	case *Wallet:
		fmt.Println("Оплата наличными?")
	case *Card:
		fmt.Println("Оплата картой?")
	case *ApplePay:
		fmt.Println("Оплата ApplePay?")
	default:
		fmt.Println("Неизвестный тип оплаты")
	}
}

func main2() {
	myWallet := &Wallet{Cash: 100}
	Buy(myWallet)

	var myMoney Payer = &Card{Balance: 100, Cardholder: "vasya"}
	Buy(myMoney)

}
