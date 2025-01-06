package main

import "fmt"

/*
В структуру можно встраивать другую вместе с её свойствами и методами.
Здесь можно вызвать метод SetName у структуры Account, хотя он есть только у Person:
acc := Account{}
acc.SetName("Name")
// то же самое будет если вызывать метод так:
// acc.Person.SetName("Name 2")
При этом для Account можно создать метод с таким же именем SetName, конфликтов не будет
То же справедливо для дублирующихся полей структур
*/
type Account struct {
	Id int
	Person
}

type Person struct {
	Name string
	Age  int
}

func (p *Person) SetName(name string) {
	p.Name = name
}

func main() {
	person := Person{}
	// person.SetName("Alex")
	person.SetName("Alexandr")

	acc := Account{1, person}
	acc.Person.SetName("Alexandr 2")

	fmt.Printf("%#v\n", acc)
}
