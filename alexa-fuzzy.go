package main

import (
//	"bytes"
	"encoding/json"
	"fmt"
	"github.com/swisskid/go-insteon/insteon"
	"io/ioutil"
//	"log"
	"net/http"
	"github.com/jeffsmith82/gofuzzy"
//	"./secrets"
	"strconv"
	"time"
)

var Accounts map[string]Acc_Info
var AccDevs map[string][]insteon.Device
var AccScenes map[string][]insteon.Scene
var AccRooms map[string][]insteon.Room
var FuzzNames map[string][]FuzzyName

func main() {
	Accounts = make(map[string]Acc_Info)
	AccDevs = make(map[string][]insteon.Device)
	AccScenes = make(map[string][]insteon.Scene)
	AccRooms = make(map[string][]insteon.Room)
	FuzzNames = make(map[string][]FuzzyName)
	http.HandleFunc("/", foo)
	http.ListenAndServe(":9003", nil)
}

func MakeFuzzy(userid string){
    for i, x := range AccScenes[userid] {
	var n FuzzyName
	n.Fuzz, _ = gofuzzy.Soundex(x.SceneName)
	n.Kind = "scene"
	n.Place = i
	FuzzNames[userid] = append(FuzzNames[userid], n)
    }
    for i, x := range AccDevs[userid] {
	var n FuzzyName
	n.Fuzz, _ = gofuzzy.Soundex(x.DeviceName)
	n.Kind = "device"
	n.Place = i
	FuzzNames[userid] = append(FuzzNames[userid], n)
    }
    for i, x := range AccRooms[userid] {
	var n FuzzyName
	n.Fuzz, _ = gofuzzy.Soundex(x.RoomName)
	n.Kind = "room"
	n.Place = i
	FuzzNames[userid] = append(FuzzNames[userid], n)
    }
    fmt.Println(FuzzNames[userid])
}

type FuzzyName struct {
    Fuzz	string
    Kind	string
    Place	int
}

type Acc_Info struct {
	Amazon       string
	Time         time.Time
}

//Request Fields
type RequestMaster struct {
	Version string  `json:"version"`
	Session Session `json:"session"`
	Request Request `json:"request"`
}
type Session struct {
	New         bool              `json:"new"`
	SessionId   string            `json:"sessionId"`
	Attributes  map[string]string `json:"attributes"`
	Application Application       `json:"application"`
	User        User              `json:"user"`
}

type Application struct {
	ApplicationId string
}
type User struct {
	UserId string
	AccessToken string
}
type Request struct { //I have to combine all the different requests into this. ugh.
	Type      string `json:"type"` //LaunchRequest,IntentRequest,SessionEndedRequest
	Timestamp string `json:"date"`
	RequestId string `json:"requestId"`
	Intent    Intent `json:"intent"`
	Reason    string `json:"reason"`
}
type Intent struct {
	Name  string          `json:"name"`
	Slots map[string]Slot `json:"slots"`
}
type Slot struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}


//Response Fields
type ResponseMaster struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
	Response          Response               `json:"response"`
}

type Response struct {
	OutputSpeech     *OutputSpeech `json:"outputSpeech,omitempty"`
	Card             *Card         `json:"card,omitempty"`
	Reprompt         *Reprompt     `json:"reprompt,omitempty"`
	ShouldEndSession bool          `json:"shouldEndSession"`
}
type OutputSpeech struct {
	Type string `json:"type,omitempty"` //must omit empty so it doesn't print the whole object
	Text string `json:"text,omitempty"` //must omit empty so it doesn't print the whole object
}
type Card struct {
	Type    string `json:"type,omitempty"` //must omit empty so it doesn't print the whole object
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}
type Reprompt struct {
	OutputSpeech OutputSpeech `json:"outputSpeech,omitempty"`
}

func foo(w http.ResponseWriter, r *http.Request) {
	var l RequestMaster
	var j ResponseMaster
	var k OutputSpeech
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Println(string(body[:]))
	json.Unmarshal(body, &l)
	// For all responses
	j.Version = "1.0"
	k.Type = "PlainText"
	j.Response.ShouldEndSession = true
	userid := l.Session.User.UserId
	accessToken := l.Session.User.AccessToken
	insteon.Access_Token = accessToken
	//client_id := insteon.Client_Id
	if l.Request.Intent.Name != "" {
		fmt.Println("The name of the intent is " + l.Request.Intent.Name)
		fmt.Println(l.Request.Intent.Slots)
		for _, slot := range l.Request.Intent.Slots {
			fmt.Println(slot.Name)
			fmt.Println(slot.Value)
		}
	}
	switch l.Request.Intent.Name {
	case "Lighting", "Activate", "Deactivate":
		fmt.Println(Accounts)
		fmt.Println("PreIF")
		fmt.Println("Gonna make fuzzy")
		MakeFuzzy(userid)
		if accessToken != "" {
			fmt.Println("Access Token Present")
			if (time.Since(Accounts[userid].Time) > 1 * time.Hour){
			    fmt.Println("Repopulating")
			    insteon.PopulateAll()
			    AccDevs[userid] = insteon.DevList
			    AccScenes[userid] = insteon.SceneList
			    AccRooms[userid] = insteon.RoomList
			    var tmpacc Acc_Info
			    tmpacc.Amazon=userid
			    tmpacc.Time=time.Now()
			    Accounts[userid] = tmpacc
			    MakeFuzzy(userid)
			}

			insteon.DevList = AccDevs[userid]
			insteon.SceneList = AccScenes[userid]
			insteon.RoomList = AccRooms[userid]
			fmt.Println(insteon.Access_Token)
		} else {
			fmt.Println("I'm going into the ELSE")
		}
		switch l.Request.Intent.Name {
		case "Lighting":
			var direction string
			var item string
			var n insteon.Command
			for _, slot := range l.Request.Intent.Slots {
				if slot.Name == "Direction" {
					direction = slot.Value
					continue
				}
				if slot.Name == "Device" {
					item = slot.Value
				}
			}
			if direction == "" {
			    direction = "on"
			}
			n.Command = direction
			item_type, id, location := insteon.SearchString(item)
			if item_type == "not_found" {
			    s, _ := gofuzzy.Soundex(item)
			    for _, x := range FuzzNames[userid]{ 
				if x.Fuzz == s {
				    item_type = x.Kind
				    location = x.Place
				    switch item_type {
					case"device":
						id = AccDevs[userid][location].DeviceID
					case"scene":
						id = AccScenes[userid][location].SceneID
					case"room":
						id = AccRooms[userid][location].RoomID
				    }
				    fmt.Println("Found with fuzzy")
				    break
				}
			    }
			}
			fmt.Println(item_type + " ID = " + strconv.Itoa(id) + " Location " + strconv.Itoa(location))
			switch item_type {
			case "device":
				n.Device_Id = id
				n.Level = insteon.DevList[location].DimLevel / 254 * 100
				k.Text = "Turning " + insteon.DevList[location].DeviceName + " " + direction
			case "scene":
				n.Scene_Id = id
				k.Text = "Turning " + insteon.SceneList[location].SceneName + " " + direction
			case "room":
				fmt.Println("Rooms are a pain")
				k.Text = "Rooms are a pain, do it yourself"
			case "not_found":
				k.Text = "The following device, scene, or room was not found: " + item
			}
			go insteon.RunCommand(n) //Forking since i don't error check
		case "Activate":
			fmt.Println(insteon.Access_Token + " is the accesss token")
			var n insteon.Command
			n.Command = "on"
			scene_name := l.Request.Intent.Slots["Scene"].Value
			ret_type, scene_id, _ := insteon.SearchString(scene_name)
			if ret_type == "scene" {
				n.Scene_Id = scene_id
				go insteon.RunCommand(n) // Forking it because I am not error checking anyway
				k.Text = "Turning the scene " + scene_name + " on"
			} else {
				k.Text = "Failed to find scene " + scene_name
			}
		case "Deactivate":
			fmt.Println(insteon.Access_Token + " is the accesss token")
			var n insteon.Command
			n.Command = "off"
			scene_name := l.Request.Intent.Slots["Scene"].Value
			ret_type, scene_id, _ := insteon.SearchString(scene_name)
			if ret_type == "scene" {
				n.Scene_Id = scene_id
				go insteon.RunCommand(n) // Forking it because I am not error checking anyway
				k.Text = "Turning the scene " + scene_name + " on"
			} else {
				k.Text = "Failed to find scene " + scene_name
			}
		}
	}
	j.Response.OutputSpeech = &k
	b, _ := json.Marshal(j)
	fmt.Println(string(b[:]))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("charset", "UTF-8")
	w.Write(b)
}
