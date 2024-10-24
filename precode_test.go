package main

import (
	"context"
	"testing"
	"time"
)

func TestGenerator(t *testing.T) {
	const count = 10
	ch := make(chan int64)
	var sum int64

	fn := func(i int64) {
		sum += i
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go Generator(ctx, ch, fn)

	results := make([]int64, 0, count)

	for i := int64(1); i <= count; i++ {
		select {
		case num := <-ch:
			results = append(results, num)
		case <-ctx.Done():
			t.Fatal("test timed out")
		}
	}

	for i := int64(1); i <= count; i++ {
		if results[i-1] != i {
			t.Errorf("expected %d, got %d", i, results[i-1])
		}
	}

	expectedSum := int64(count * (count + 1) / 2)
	if sum != expectedSum {
		t.Errorf("expected sum %d, got %d", expectedSum, sum)
	}
}

func TestGeneratorWithContextCancellation(t *testing.T) {
	ch := make(chan int64)

	ctx, cancel := context.WithCancel(context.Background())
	close(ch)
	cancel()

	var sum int64
	fn := func(i int64) {
		sum += i
	}

	go Generator(ctx, ch, fn)

	time.Sleep(10 * time.Millisecond)

	if sum != 0 {
		t.Errorf("expected sum to be 0, got %d", sum)
	}
}
