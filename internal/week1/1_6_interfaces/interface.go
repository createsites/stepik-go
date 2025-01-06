package main

import "fmt"

type Payer interface {
	Pay(int) error
}

type Wallet struct {
	Balance int
}

func (w *Wallet) Pay(price int) error {
	if w.Balance < price {
		return fmt.Errorf("no fonds")
	}
	w.Balance -= price

	return nil
}

func Buy(p Payer, price int) {
	p.Pay(price)
	// мы не можем обратиться к полю структуры, скрывающейся за интерфейсом, напрямую вот так
	// p.Balance
	// но можно вот так
	// p.(*Wallet).Balance
	// т.е. конкретизировав реализацию интерфейса (Wallet)
	// конкретную реализацию можно поймать, перечисляя структуры в switch case
}

func main() {
	// здесь в переменную типа интерфейс обязательно передавать ссылку на структуру
	// просто Wallet{Balance: 100} будет ошибка
	var wal Payer = &Wallet{Balance: 100}
	Buy(wal, 20)
	fmt.Printf("%#v\n", wal)
}