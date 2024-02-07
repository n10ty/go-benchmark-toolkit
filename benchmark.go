package benchmark

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync/atomic"
	"time"

	"github.com/schollz/progressbar/v3"
)

type Result struct {
	sum        atomic.Int64
	mins       map[int]time.Duration
	maxs       map[int]time.Duration
	min        time.Duration
	max        time.Duration
	times      []time.Duration
	success    atomic.Int64
	failed     atomic.Int64
	median     time.Duration
	total      time.Duration
	iterations atomic.Int64
	p80        time.Duration
	p95        time.Duration
	p99        time.Duration
}

type Benchmark struct {
	executable func() error
	threads    int
	result     Result
	duration   time.Duration
}

func NewBenchmark(duration time.Duration, threads int, executable func() error) *Benchmark {
	b := &Benchmark{
		executable: executable,
		threads:    threads,
		duration:   duration,
		result: Result{
			times: make([]time.Duration, 0),
			mins:  make(map[int]time.Duration),
			maxs:  make(map[int]time.Duration),
			min:   time.Duration(math.MaxInt64),
			max:   time.Duration(math.MinInt64),
		},
	}
	for i := 0; i < threads; i++ {
		b.result.mins[i] = time.Duration(math.MaxInt64)
		b.result.maxs[i] = time.Duration(math.MinInt64)
	}
	return b
}

func (g *Benchmark) launch(executable func() error, i int, bar *progressbar.ProgressBar, stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
			start := time.Now()
			err := executable()
			if err != nil {
				g.result.failed.Add(1)
			} else {
				g.result.success.Add(1)
			}
			elapsed := time.Since(start)

			g.result.sum.Add(int64(elapsed))
			g.result.times = append(g.result.times, elapsed)

			if elapsed < g.result.mins[i] {
				g.result.mins[i] = elapsed
			}

			if elapsed > g.result.maxs[i] {
				g.result.maxs[i] = elapsed
			}
			g.result.total += elapsed
			g.result.iterations.Add(1)
			bar.Add(1)
		}
	}
}

func (g *Benchmark) Run() {
	bar := progressbar.Default(-1)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(g.duration))
	defer cancel()

	for i := 0; i < g.threads; i++ {
		go g.launch(g.executable, i, bar, ctx.Done())
	}

	<-ctx.Done()
	for t := 0; t < g.threads; t++ {
		sort.Slice(g.result.times, func(i, j int) bool {
			return g.result.times[i] < g.result.times[j]
		})
	}

	if g.threads%2 == 0 {
		g.result.median = (g.result.times[g.threads/2-1] + g.result.times[g.threads/2]) / 2
	} else {
		g.result.median = g.result.times[g.threads/2]
	}

	p80 := float64(len(g.result.times)) * 0.8
	g.result.p80 = g.result.times[int(p80)]
	p95 := float64(len(g.result.times)) * 0.95
	g.result.p95 = g.result.times[int(p95)]
	p99 := float64(len(g.result.times)) * 0.99
	g.result.p99 = g.result.times[int(p99)]

	for i := 0; i < g.threads; i++ {
		if g.result.mins[i] < g.result.min {
			g.result.min = g.result.mins[i]
		}
		if g.result.maxs[i] > g.result.max {
			g.result.max = g.result.maxs[i]
		}
	}
}

func (g *Benchmark) PrintSummary() {
	fmt.Println()
	fmt.Println("Avg time:", time.Duration(g.result.sum.Load()/g.result.iterations.Load()))
	fmt.Println("Min time:", g.result.min)
	fmt.Println("Max time:", g.result.max)
	fmt.Println("P95:", g.result.p95)
	fmt.Println("P99:", g.result.p99)
	fmt.Println("Median time:", g.result.median)
	fmt.Println("Total time:", g.result.total)
	fmt.Println("Success:", g.result.success.Load())
	fmt.Println("Failed:", g.result.failed.Load())
	fmt.Println("RPS:", float64(g.result.iterations.Load())/g.result.total.Seconds())
	fmt.Println("Iterations:", g.result.iterations.Load())
}
