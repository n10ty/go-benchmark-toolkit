# Simple toolkit for performance testing in GO

## Usage
```go get github.com/n10ty/go-benchmark-toolkit```

## Example
```go
func main() {
	exec := func() error {
		time.Sleep(time.Millisecond * 100)
		return nil
	}

	threads := 10
	b := NewBenchmark(time.Second*10, threads, exec)
	b.Run()
	b.PrintSummary()
	
	//Output:
	//Avg time: 100.97613ms
	//Min time: 100.104542ms
	//Max time: 101.072958ms
	//Median time: 100.47225ms
	//Total time: 9.895660752s
	//Iterations: 98
}
```
