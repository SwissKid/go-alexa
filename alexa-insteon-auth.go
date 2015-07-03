package main

import (
	//      "io"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/swisskid/go-insteon/insteon"
	"io/ioutil"
	"log"
	"net/http"
	//"net/url"
	//"strings"
	"./secrets"
	"strconv"
)

var Accounts map[string]Acc_Info
var AccDevs map[string][]insteon.Device
var AccScenes map[string][]insteon.Scene
var AccRooms map[string][]insteon.Room

func main() {
	Accounts = make(map[string]Acc_Info)
	AccDevs = make(map[string][]insteon.Device)
	AccScenes = make(map[string][]insteon.Scene)
	AccRooms = make(map[string][]insteon.Room)
	http.HandleFunc("/", foo)
	http.ListenAndServe(":9003", nil)
}

type Acc_Info struct {
	Amazon       string
	Access_Token string
	Refresh      string
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

var Account_Location = "/srv/alexa/accounts/"

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
	client_id := insteon.Client_Id
	if l.Request.Intent.Name != "" {
		fmt.Println("The name of the intent is " + l.Request.Intent.Name)
		fmt.Println(l.Request.Intent.Slots)
		for _, slot := range l.Request.Intent.Slots {
			fmt.Println(slot.Name)
			fmt.Println(slot.Value)
		}
	}
	switch l.Request.Intent.Name {
	case "Register":
		var m Card
		long_url := "http://connect.insteon.com/api/v2/oauth2/auth?client_id=" + client_id + "&state=" + userid + "&response_type=code&redirect_uri=http://veryoblivio.us:9001/"
		short_url := ShortenURL(long_url)
		m.Type = "Simple"
		m.Title = "Register your Insteon Connection"
		m.Content = short_url
		j.Response.Card = &m
		k.Text = "Please follow the link on the card to sign into your insteon account"
		break
	case "Lighting", "Activate":
		fmt.Println(Accounts)
		fmt.Println("PreIF")
		if val, ok := Accounts[userid]; ok {
			fmt.Println("I'm going into the IF")
			insteon.Access_Token = val.Access_Token
			insteon.DevList = AccDevs[userid]
			insteon.SceneList = AccScenes[userid]
			insteon.RoomList = AccRooms[userid]
			fmt.Println("Post assignment")
			fmt.Println(insteon.Access_Token)
		} else {
			fmt.Println("I'm going into the ELSE")
			if val2, er2 := GetAccInfo(userid); er2 { //er2 is actually success if false - like an error
				fmt.Println("I'm going into the ELSEELSE")
				var m Card
				client_id := insteon.Client_Id
				long_url := "http://connect.insteon.com/api/v2/oauth2/auth?client_id=" + client_id + "&state=" + userid + "&response_type=code&redirect_uri=http://veryoblivio.us:9001/"
				short_url := ShortenURL(long_url)
				m.Type = "Simple"
				m.Title = "Register your Insteon Connection"
				m.Content = short_url
				j.Response.Card = &m
				k.Text = "No linked account found. Please follow the link on the card to sign into your insteon account"
			} else {
				fmt.Println("I'm going into the ELSEIF")
				insteon.Access_Token = val2.Access_Token
				insteon.PopulateAll()
				AccDevs[userid] = insteon.DevList
				AccScenes[userid] = insteon.SceneList
				AccRooms[userid] = insteon.RoomList
				Accounts[userid] = val2
				fmt.Println("Post assignment")
				fmt.Println(insteon.Access_Token)
				fmt.Println(Accounts)
			}
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
			n.Command = direction
			item_type, id, location := insteon.SearchString(item)
			fmt.Println(item_type + " ID = " + strconv.Itoa(id) + " Location " + strconv.Itoa(location))
			switch item_type {
			case "device":
				n.Device_Id = id
				n.Level = insteon.DevList[location].DimLevel / 254 * 100
				k.Text = "Turning " + item + " " + direction
			case "scene":
				n.Scene_Id = id
				k.Text = "Turning " + item + " " + direction
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
		}
	}
	j.Response.OutputSpeech = &k
	b, _ := json.Marshal(j)
	fmt.Println(string(b[:]))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("charset", "UTF-8")
	w.Write(b)
}

func GetAccInfo(amazon_id string) (account_info Acc_Info, success bool) {
	account_info.Amazon = amazon_id
	body, err := ioutil.ReadFile(Account_Location + amazon_id)
	if err != nil {
		success = false
		return
	}
	fmt.Println("Got token from file for " + amazon_id)
	account_info.Refresh = string(body)
	account_info.Access_Token, success = insteon.Refresh_Bearer(string(body))
	fmt.Println("Got refresh " + amazon_id + " and its " + account_info.Access_Token)
	if Accounts == nil {
		Accounts = make(map[string]Acc_Info)
	}
	Accounts[amazon_id] = account_info
	return
}

type toShorten struct {
	LongUrl string `json:"longUrl"`
}

func ShortenURL(long_url string) (shorturl string) {
	u := "https://www.googleapis.com/urlshortener/v1/url" + "?key=" + secrets.Google_api_key
	var d toShorten
	d.LongUrl = long_url
	data, _ := json.Marshal(d)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", u, bytes.NewBuffer(data))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Errored when sending request to the server")
		return
	}
	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)
	var i map[string]string
	json.Unmarshal(resp_body, &i)
	shorturl = i["id"]
	return
}
