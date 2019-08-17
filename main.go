package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync/atomic"
	"syscall"
	"time"
)

type Chars []rune

func (c Chars) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(c))
}

func (c *Chars) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, c)
}

type Word struct {
	Chars Chars
	Path  Path
}

type WordSet []Word

// Change to int16 if you want boards bigger than 10x10.
type Unit int8

type Point struct {
	X Unit
	Y Unit
}

func (p Point) MarshalJSON() ([]byte, error) {
	return json.Marshal([]Unit{p.X, p.Y})
}

func (p *Point) UnmarshalJSON(data []byte) error {
	var v []Unit
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	p.X = v[0]
	p.Y = v[1]

	return nil
}

type Rect struct {
	Min, Max Point
}

type Path []Point

func (a Path) Index(p Point) int {
	for i := range a {
		if a[i] == p {
			return i
		}
	}

	return -1
}

func (a Path) Same(b Path) bool {
	if len(a) != len(b) {
		return false
	}

	for _, p := range a {
		if b.Index(p) == -1 {
			return false
		}
	}

	return true
}

var (
	boardWidth  = flag.Int("w", 5, "Board's width.")
	boardHeight = flag.Int("h", 5, "Board's height.")
	validate    = flag.String("validate", "", "Paste a dump to validate it.")
	cpuprofile  = flag.String("cpuprofile", "", "write cpu profile to file")
)

type Result struct {
	Board   *Board
	WordSet WordSet
}

func main() {
	rand.Seed(time.Now().UnixNano())

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *validate != "" {
		var result Result
		if err := json.Unmarshal([]byte(*validate), &result); err != nil {
			log.Fatal(err)
		}

		v := NewValidator(result.Board)
		ok := v.Validate(result.WordSet)

		if ok {
			fmt.Println("Valid!")
		} else {
			fmt.Println("Not valid yet :(")
		}

		return
	}

	totalLen := 0
	words := make([]Chars, len(flag.Args()))
	for i, w := range flag.Args() {
		chars := Chars(w)
		words[i] = chars
		totalLen += len(chars)
	}
	if totalLen > (*boardWidth)*(*boardHeight) {
		log.Fatalf("Total length of words (%d) is bigger than the boards capacity (%d).",
			totalLen, (*boardWidth)*(*boardHeight))
	}

	// Some random noise to fill the board with.
	// noise := make(Chars, 5*5)
	// for i := range noise {
	// 	noise[i] = rune(rand.Intn('ת'-'א') + 'א')
	// }

	// Attempt to generate a valid board concurrently.
	start := time.Now()
	var tries uint64
	var validations uint64

	winner := make(chan Result)

	for i := 0; i < runtime.NumCPU()*2; i++ {
		go func() {
			b := NewBoard(Unit(*boardWidth), Unit(*boardHeight))
			filler := NewFiller(b)

			for {
				atomic.AddUint64(&tries, 1)

				// b.Fill(noise)

				ws, ok := filler.Fill(words)
				if !ok {
					b.Reset()
					filler.Reset()
					continue
				}

				validator := NewValidator(b)
				atomic.AddUint64(&validations, 1)
				if !validator.Validate(ws) {
					b.Reset()
					filler.Reset()
					continue
				}

				winner <- Result{
					Board:   b,
					WordSet: ws,
				}
				return
			}
		}()
	}

	// Wait for a valid board from one of the workers,
	// while reporting status every few seconds.
	var result Result
	var prevTries uint64

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

Loop:
	for {
		select {
		case result = <-winner:
			break Loop
		case <-time.After(time.Second * 1):
			fmt.Printf("\rIteration #%.2fM [%dK TP/s] (%.2fM (%.2f%%) validations)",
				float64(tries)/1e6,
				(tries-prevTries)/1e3/1,
				float64(validations)/1e6,
				float64(validations)/float64(tries)*100,
			)
			prevTries = tries
		case <-interrupt:
			fmt.Println()
			fmt.Printf("Stopping after %.2fM tries...", float64(tries)/1e6)
			return
		}
	}

	fmt.Println()
	fmt.Println()
	fmt.Printf("Completed after %d tries with %d validations. Took %s\n", tries, validations, time.Since(start))
	fmt.Println()

	b, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Dump:\n%s\n", string(b))
	fmt.Println()

	fmt.Println("Board:")
	fmt.Println()
	result.Board.Render(os.Stdout)
}
