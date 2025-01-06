/*
Объяснение работы замыканий.
Замыкание - функция, которая использует переменные извне.
В примере ниже мы определяем тип closureFunc, чтобы вернуть его как функцию-замыкание
closureFunc - принимает строку и ничего не возвращает
Функция prefixer принимает строку prefix и возвращает функцию-замыкание, которая принимает строку in и выводит её с префиксом
Переменная successLogger является замыканием, далее она вызывается и выводит строку с указанным префиксом
*/
package main

import "fmt"

/*
Объявляем тип для возвращаемой функции в prefixer.
Можно было бы обойтись без этого, тогда объявление prefixer было бы такое:
prefixer := func(prefix string) func(string) {
*/
type closureFunc func(string)

func main() {
	//leetcode.DuplicateZeros()
	prefixer := func(prefix string) closureFunc {
		return func(in string) {
			fmt.Printf("[%s], %s\n", prefix, in)
		}
	}
	// Создаем замыкание с префиксом "SUCCESS"
	successLogger := prefixer("SUCCESS")
	// Вызываем замыкание
	successLogger("expected behaviour")
}

/*
// Вот еще хороший пример со скидками

func discount(discountPercentage float64) func(float64) float64 {
    return func(price float64) float64 {
        return price - (price * discountPercentage / 100)
    }
}

func main() {
    discount10 := discount(10)
    discount25 := discount(25)

    fmt.Println(discount10(100)) // 90
    fmt.Println(discount25(100)) // 75
}
*/