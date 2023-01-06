package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
	"strings"
    "strconv"
)

func main() {
    // open file
    fileName := os.Args[1]
    f, err := os.Open("traces/" + fileName)
    if err != nil {
        log.Fatal(err)
    }
    // remember to close the file at the end of the program
    defer f.Close()
    sizeS := os.Args[2]
    pagesS := os.Args[3]
    size, _ := strconv.Atoi(sizeS)
    pages, _ := strconv.Atoi(pagesS)
    cache := NewARC(int(size), int(pages))

    // read the file line by line using scanner
    scanner := bufio.NewScanner(f)

    for scanner.Scan() {
        line := fmt.Sprintf("%s", scanner.Text())
		split := strings.Fields(line)
		_, ok := cache.Get(split[1])
		if !ok{
			cache.Set(split[1], []byte(split[1]))
		}
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }

	fmt.Print("Hits: ")
	fmt.Println(cache.hits)
	fmt.Print("Misses: ")
	fmt.Println(cache.misses)
    fmt.Print("Ratio: ")
	fmt.Println(float64(cache.hits)/(float64(cache.misses)+float64(cache.hits)))
}