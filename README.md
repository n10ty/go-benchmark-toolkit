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

	threads := 20
	b := NewBenchmark(time.Second*10, threads, exec)
	b.Run()
	b.PrintSummary()
	
	//Output:
	//Avg time: 173.865291ms
    //Min time: 115.557292ms
    //Max time: 405.42125ms
    //Median time: 115.557292ms
    //Total time: 9.910321625s
    //Success: 57
    //Failed: 0
    //RPS: 5.751579227884039
    //Iterations: 57
}
```
