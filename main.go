package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"
    "github.com/go-martini/martini"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"github.com/martini-contrib/render"
	"github.com/gorilla/websocket"
	"dices"
)

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
