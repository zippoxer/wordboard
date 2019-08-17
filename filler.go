package main

import (
	"math/rand"
	"time"

	"github.com/dgryski/go-pcgr"
)

const (
	MaxWords       = 100
	MaxWordLength  = 20
	MaxNearbySlots = 9
)

type Filler struct {
	b *Board

	rand *rand.Rand

	// A slice the same is the same size as the board
	// but instead of storing characters at each slot,
	// it stores whether the slot is used or not.
	totalUsed int

	// Working memory for randPerm.
	perm []int

	// Working memory for unusedPoints.
	unused []Point

	// Working memory for word path.
	path Path

	// Working memory for wordSet.
	wordSet WordSet
}

func NewFiller(b *Board) *Filler {
	rngSource := pcgr.New(time.Now().UnixNano(), 0)

	return &Filler{
		b:    b,
		rand: rand.New(&rngSource),
		// rand: rand.New(rand.NewSource(time.Now().UnixNano())),
		totalUsed: 0,
		perm:      make([]int, MaxWords),
		unused:    make([]Point, 0, MaxNearbySlots),
		path:      make(Path, MaxWordLength),
		wordSet:   make(WordSet, 0, MaxWords),
	}
}

func (f *Filler) Reset() {
	f.totalUsed = 0
}

func (f *Filler) Fill(words []Chars) (WordSet, bool) {
	set := f.wordSet[:0]

	for _, i := range f.randPerm(len(words)) {
		path, ok := f.add(words[i])
		if !ok {
			return nil, false
		}

		set = append(set, Word{
			Chars: words[i],
			Path:  path,
		})
	}

	setCopy := make(WordSet, len(set))
	copy(setCopy, set)
	return setCopy, true
}

func (f *Filler) add(word Chars) (Path, bool) {
	// Get a random unused point to insert the point at.

	// 1. The Easy Way:
	// p, ok := f.randUnusedPoint(Rect{Point{0, 0}, Point{f.b.Width - 1, f.b.Height - 1}})
	// if !ok {
	// 	return nil, false
	// }

	// 2. The Fast Way:
	//
	// Pick a random index of an unused slot.
	totalUnused := int(f.b.Width)*int(f.b.Height) - f.totalUsed
	if totalUnused == 0 {
		return nil, false
	}
	randUnusedIndex := int(f.rand.Int31n(int32(totalUnused)))

	// Find the index of that slot on the board.
	var (
		p           Point
		ok          bool
		unusedIndex = 0
	)
Loop:
	for x := Unit(0); x < f.b.Width; x++ {
		for y := Unit(0); y < f.b.Height; y++ {
			if f.b.Data[x][y] == 0 {
				if unusedIndex == randUnusedIndex {
					p = Point{x, y}
					ok = true
					break Loop
				}

				unusedIndex++
			}
		}
	}
	if !ok {
		panic("Can't find unused slot, even though there is.")
	}

	// Attempt to insert the word into the board at the chosen slot.
	path := f.path[:0]
	for i, r := range word {
		f.set(p, r)
		path = append(path, p)

		if i == len(word)-1 {
			continue
		}

		r := Rect{
			Point{
				X: p.X - 1,
				Y: p.Y - 1,
			},
			Point{
				X: p.X + 1,
				Y: p.Y + 1,
			},
		}
		if r.Min.X < 0 {
			r.Min.X = 0
		}
		if r.Min.Y < 0 {
			r.Min.Y = 0
		}
		if r.Max.X > f.b.Width-1 {
			r.Max.X = f.b.Width - 1
		}
		if r.Max.Y > f.b.Height-1 {
			r.Max.Y = f.b.Height - 1
		}

		p, ok = f.randUnusedPoint(r)
		if !ok {
			return nil, false
		}
	}

	pathCopy := make(Path, len(path))
	copy(pathCopy, path)

	return pathCopy, true
}

func (f *Filler) set(p Point, r rune) {
	f.totalUsed++
	f.b.Set(p, r)
}

func (f *Filler) randUnusedPoint(r Rect) (Point, bool) {
	// Find all unused points within the given bounds.
	points := f.unused[:0]
	for x := r.Min.X; x <= r.Max.X; x++ {
		for y := r.Min.Y; y <= r.Max.Y; y++ {
			if f.b.Data[x][y] == 0 {
				points = append(points, Point{x, y})
			}
		}
	}

	if len(points) == 0 {
		return Point{}, false
	}

	// Choose and return a random point.
	p := points[f.rand.Int31n(int32(len(points)))]
	return p, true
}

// Same as math/rand.Perm but avoids slice re-allocation.
func (f *Filler) randPerm(n int) []int {
	if len(f.perm) < n {
		f.perm = make([]int, n)
	}

	for i := 0; i < n; i++ {
		j := f.rand.Intn(i + 1)
		f.perm[i] = f.perm[j]
		f.perm[j] = i
	}

	return f.perm[:n]
}
