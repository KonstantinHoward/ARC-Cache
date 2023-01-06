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
    lru := NewLru(int(size), int(pages))

    // read the file line by line using scanner
    scanner := bufio.NewScanner(f)

    for scanner.Scan() {
        line := fmt.Sprintf("%s", scanner.Text())
		split := strings.Fields(line)
		_, ok := lru.Get(split[1])
		if !ok{
			lru.Set(split[1], []byte(split[1]))
		}
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }

	fmt.Print("Hits: ")
	fmt.Println(lru.stat.Hits)
	fmt.Print("Misses: ")
	fmt.Println(lru.stat.Misses)
	fmt.Print("Ratio: ")
	fmt.Println(float64(lru.stat.Hits)/(float64(lru.stat.Misses)+float64(lru.stat.Hits)))
}
