package benchmark

import (
	"context"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"math"
	"sort"
	"sync"
	"time"
)

type Result struct {
	sum        time.Duration
	min        time.Duration
	max        time.Duration
	times      []time.Duration
	success    int64
	failed     int64
	median     time.Duration
	total      time.Duration
	iterations int64
}

type Benchmark struct {
	executable func() error
	threads    int
	result     Result
	duration   time.Duration
	m          sync.Mutex
}

func NewBenchmark(duration time.Duration, threads int, executable func() error) *Benchmark {
	return &Benchmark{
		executable: executable,
		threads:    threads,
		result: Result{
			min: time.Duration(math.MaxInt64),
			max: time.Duration(math.MinInt64),
		},
		duration: duration,
		m:        sync.Mutex{},
	}
}

func (g *Benchmark) launch(executable func() error, bar *progressbar.ProgressBar, stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
			g.m.Lock()
			start := time.Now()
			err := executable()
			if err != nil {
				g.result.failed++
			} else {
				g.result.success++
			}
			elapsed := time.Since(start)

			g.result.sum += elapsed
			g.result.times = append(g.result.times, elapsed)

			if elapsed < g.result.min {
				g.result.min = elapsed
			}

			if elapsed > g.result.max {
				g.result.max = elapsed
			}
			g.result.total += elapsed
			g.result.iterations++
			bar.Add(1)
			g.m.Unlock()
		}
	}
}

func (g *Benchmark) Run() {
	bar := progressbar.Default(-1)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(g.duration))
	defer cancel()

	for i := 0; i < g.threads; i++ {
		go g.launch(g.executable, bar, ctx.Done())
	}

	<-ctx.Done()
	sort.Slice(g.result.times, func(i, j int) bool {
		return g.result.times[i] < g.result.times[j]
	})

	if g.threads%2 == 0 {
		g.result.median = (g.result.times[g.threads/2-1] + g.result.times[g.threads/2]) / 2
	} else {
		g.result.median = g.result.times[g.threads/2]
	}
}

func (g *Benchmark) PrintSummary() {
	fmt.Println()
	fmt.Println("Avg time:", g.result.sum/time.Duration(g.result.iterations))
	fmt.Println("Min time:", g.result.min)
	fmt.Println("Max time:", g.result.max)
	fmt.Println("Median time:", g.result.median)
	fmt.Println("Total time:", g.result.total)
	fmt.Println("Success:", g.result.success)
	fmt.Println("Failed:", g.result.failed)
	fmt.Println("Iterations:", g.result.iterations)
}
