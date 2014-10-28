package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
    "github.com/go-martini/martini"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"github.com/martini-contrib/render"
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


type RollRecord struct {
	DiceDef string
	Roll int
	Timestamp time.Time
}

func CreateRollRecord(pool *DicePool, c *mgo.Collection) {
	record := RollRecord{DiceDef: pool.DiceDef, Roll: pool.LastRoll, Timestamp: time.Now()}
	err := c.Insert(&record)
	if err != nil {
		log.Fatal(err)
	}
	return
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

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	pattern := `\dd\d+([\+-]\d+)?`
	fmt.Println("Started")
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Directory: "templates",
		Extensions: []string{".tmpl", ".html"},
	}))
	m.Use(martini.Static("static", martini.StaticOptions{Prefix: "static"}))
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("test").C("rolls")
	m.Get("/roll/:def", func(params martini.Params) (int, string) {
		match, _ := regexp.MatchString(pattern, params["def"])
		if match {
			p := CreateDicePool(params["def"])
			roll := p.Roll()
			CreateRollRecord(p, c)
			return 200, strconv.Itoa(roll)
		} else {
			return 403, "Wrong format!"
		}
	})
	m.Get("/rolls", func(r render.Render) {
		var rolls []RollRecord
		c.Find(bson.M{}).Iter().All(&rolls)
		fmt.Println(rolls)
		r.HTML(200, "rolls", rolls)
	})
	m.Run()
}
