package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"
    "github.com/go-martini/martini"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"github.com/martini-contrib/render"
	"github.com/gorilla/websocket"
	"dices"
	"net/http"
	"ws_helpers"
	"encoding/json"
)


type RollRecord struct {
	DiceDef 	string		`json:"dice_def"`
	Roll 		int			`json:"roll"`
	Timestamp 	time.Time 	`json:"timestamp"`
}

func CreateRollRecord(pool *dices.DicePool, c *mgo.Collection) *RollRecord {
	record := new(RollRecord)
	record.DiceDef = pool.DiceDef
	record.Roll = pool.LastRoll
	record.Timestamp = time.Now()
	err := c.Insert(&record)
	if err != nil {
		log.Fatal(err)
	}
	return record
}

func WSHandler(w http.ResponseWriter, r *http.Request, c *mgo.Collection) {
	log.Println(ws_helpers.ActiveClients)
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	client := ws.RemoteAddr()
	pattern := `\dd\d+([\+-]\d+)?`
	sockCli := ws_helpers.ClientConn{ws, client}
	ws_helpers.AddClient(sockCli)

	for {
		var ret []byte;
		log.Println(len(ws_helpers.ActiveClients), ws_helpers.ActiveClients)
		messageType, p, err := ws.ReadMessage()
		def := string(p)
		if err != nil {
			ws_helpers.DeleteClient(sockCli)
			log.Println("bye")
			log.Println(err)
			return
		}
		match, _ := regexp.MatchString(pattern, def)
		if match {
			p := dices.CreateDicePool(def)
			p.Roll()
			r := struct {
					Type string
					Message *RollRecord
				}{"roll", CreateRollRecord(p, c)}
			ret, err = json.Marshal(r)
			ws_helpers.BroadcastMessage(messageType, ret)
		} else {
			r := struct {
					Type string
					Message string
				}{"error", "Wrong format!"}
			ret, err = json.Marshal(r)
			if err := ws.WriteMessage(messageType, ret); err != nil {
				return
			}
		}
	}
}


func main() {
	rand.Seed(time.Now().UTC().UnixNano())
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
	m.Map(c)
	m.Get("/rolls", func(r render.Render) {
		var rolls []RollRecord
		c.Find(bson.M{}).Iter().All(&rolls)
		r.HTML(200, "rolls", rolls)
	})
	m.Get("/sock", WSHandler)
	m.Run()
}
