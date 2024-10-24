package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

func Generator(ctx context.Context, ch chan<- int64, fn func(int64)) {
	// Добавим флаг для контроля закрытия канала
	closed := false
	defer func() {
		if !closed {
			close(ch) // Закрываем канал только если он не был закрыт
			closed = true
		}
	}()

	for i := int64(1); ; i++ {
		select {
		case <-ctx.Done():
			return // Выходим, если контекст отменен
		case ch <- i:
			fn(i) // Вызываем функцию для подсчета
		}
	}
}

// Worker читает число из канала in и пишет его в канал out.
func Worker(in <-chan int64, out chan<- int64) {
	defer close(out) // Закрываем выходной канал после завершения работы
	for num := range in {
		out <- num // Передаем число дальше
	}
}

func main() {
	chIn := make(chan int64)

	// Создание контекста с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// для проверки будем считать количество и сумму отправленных чисел
	var inputSum int64
	var inputCount int64

	// генерируем числа, считая параллельно их количество и сумму
	go Generator(ctx, chIn, func(i int64) {
		inputSum += i
		inputCount++
	})

	const NumOut = 5 // количество обрабатывающих горутин и каналов
	outs := make([]chan int64, NumOut)
	for i := 0; i < NumOut; i++ {
		outs[i] = make(chan int64)
		go Worker(chIn, outs[i])
	}

	// amounts — слайс, в который собирается статистика по горутинам
	amounts := make([]int64, NumOut)
	// chOut — канал, в который будут отправляться числа из горутин outs[i]
	chOut := make(chan int64)

	var wg sync.WaitGroup

	// Собираем числа из каналов outs
	for i := 0; i < NumOut; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for num := range outs[i] {
				chOut <- num
				amounts[i]++
			}
		}(i)
	}

	// горутина для закрытия chOut после завершения всех работников
	go func() {
		wg.Wait()
		close(chOut)
	}()

	var count int64 // количество чисел результирующего канала
	var sum int64   // сумма чисел результирующего канала

	for num := range chOut {
		sum += num
		count++
	}

	fmt.Println("Количество чисел", inputCount, count)
	fmt.Println("Сумма чисел", inputSum, sum)
	fmt.Println("Разбивка по каналам", amounts)

	// Проверка результатов
	if inputSum != sum {
		log.Fatalf("Ошибка: суммы чисел не равны: %d != %d\n", inputSum, sum)
	}
	if inputCount != count {
		log.Fatalf("Ошибка: количество чисел не равно: %d != %d\n", inputCount, count)
	}
	for _, v := range amounts {
		inputCount -= v
	}
	if inputCount != 0 {
		log.Fatalf("Ошибка: разделение чисел по каналам неверное\n")
	}
}
