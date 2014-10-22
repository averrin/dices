package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Dice struct {
	Sides    int
	LastRoll int
}

type DicePool struct {
	Dices    []Dice
	LastRoll int
	Mult     int
}

func NewDice(sides int) *Dice {
	d := new(Dice)
	d.Sides = sides
	return d
}

func NewDicePool(count int, sides int, mult int) *DicePool {
	p := new(DicePool)
	for i := 0; i <= count-1; i++ {
		d := NewDice(sides)
		p.Dices = append(p.Dices, *d)
	}
	p.Mult = mult
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
	return sum + p.Mult
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
	return NewDicePool(count, sides, mult)

}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	var def string
	var match bool
	pattern := `\dd\d+([\+-]\d+)?`
	for {
		fmt.Print("Roll def: ")
		fmt.Scan(&def)
		match, _ = regexp.MatchString(pattern, def)
		for !match {
			fmt.Print("Roll def: ")
			fmt.Scan(&def)
			match, _ = regexp.MatchString(pattern, def)
		}
		p := CreateDicePool(def)
		fmt.Println(p.Roll())
	}
}
