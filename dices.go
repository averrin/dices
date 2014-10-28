package dices

import (
	"strings"
	"math/rand"
	"regexp"
	"log"
	"time"
	"strconv"
)

type Dice struct {
	Sides    int
	LastRoll int
}

type DicePool struct {
	Dices    []Dice
	LastRoll int
	Mult     int
	DiceDef string
}

func NewDice(sides int) *Dice {
	d := new(Dice)
	d.Sides = sides
	return d
}

func NewDicePool(count int, sides int, mult int, def string) *DicePool {
	p := new(DicePool)
	for i := 0; i <= count-1; i++ {
		d := NewDice(sides)
		p.Dices = append(p.Dices, *d)
	}
	p.Mult = mult
	p.DiceDef = def
	return p
}

func (d *Dice) Roll() int {
	r := rand.Intn(d.Sides) + 1
	d.LastRoll = r
	return r
}


func (p *DicePool) Roll() int {
	sum := 0
	for _, d := range p.Dices {
		sum += d.Roll()
	}
	sum += p.Mult
	p.LastRoll = sum
	return sum
}

func CreateDicePool(def string) *DicePool {
	var sides int
	parsed_def := strings.Split(def, "d")
	count, _ := strconv.Atoi(parsed_def[0])
	pattern := `(\d+)([\+-])?(\d+)?`
	re := regexp.MustCompile(pattern)
	mult := 0
	m, _ := regexp.MatchString(pattern, parsed_def[1])
	if !m {
		sides, _ = strconv.Atoi(parsed_def[1])
	} else {
		part2 := re.FindAllStringSubmatch(parsed_def[1], -1)
		sides, _ = strconv.Atoi(part2[0][1])
		mult, _ = strconv.Atoi(part2[0][2] + part2[0][3])
	}
	return NewDicePool(count, sides, mult, def)

}
