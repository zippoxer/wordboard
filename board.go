package main

import (
	"encoding/json"
	"io"

	"github.com/olekukonko/tablewriter"
)

const (
	MaxWidth  = 10
	MaxHeight = 10
)

type BoardData [MaxWidth][MaxHeight]rune

func (d BoardData) MarshalJSON() ([]byte, error) {
	var v [MaxWidth][MaxHeight]string
	for x := 0; x < MaxWidth; x++ {
		for y := 0; y < MaxHeight; y++ {
			v[x][y] = string(d[x][y])
		}
	}

	return json.Marshal(&v)
}
func (d *BoardData) UnmarshalJSON(data []byte) error {
	var v [MaxWidth][MaxHeight]string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	for x := 0; x < MaxWidth; x++ {
		for y := 0; y < MaxHeight; y++ {
			d[x][y] = rune(v[x][y][0])
		}
	}
	return nil
}

type Board struct {
	Width  Unit
	Height Unit
	Data   BoardData
}

func NewBoard(width, height Unit) *Board {
	return &Board{
		Width:  width,
		Height: height,
		Data:   BoardData{},
	}
}

func (b *Board) At(p Point) rune {
	return b.Data[p.X][p.Y]
}

func (b *Board) Set(p Point, char rune) {
	b.Data[p.X][p.Y] = char
}

func (b *Board) Reset() {
	b.Data = BoardData{}
}

func (b *Board) Render(w io.Writer) {
	table := tablewriter.NewWriter(w)

	for y := Unit(0); y < b.Height; y++ {
		row := make([]string, b.Width)
		for x := Unit(0); x < b.Width; x++ {
			c := b.At(Point{x, y})
			if c == 0 {
				c = '-'
			}
			row[x] = string(c)
		}
		table.Append(row)
	}

	table.Render()
}
