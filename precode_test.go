package main

import (
	"context"
	"testing"
	"time"
)

func TestGeneratorWithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan int64)
	go Generator(ctx, ch, func(n int64) {})

	sum := int64(0)

	for i := 0; i < 10; i++ {
		sum += <-ch
	}

	cancel() // Отмена контекста для завершения генератора

	// Даем генератору некоторое время, чтобы завершить работу
	time.Sleep(100 * time.Millisecond)

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed, but it is still open")
		}
	default:
		// Не надо ничего делать, все хорошо
	}

	if sum != 55 {
		t.Errorf("expected sum 55, got %d", sum)
	}
}
