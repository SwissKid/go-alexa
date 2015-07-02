package main

import (
        //      "io"
        "encoding/json"
        "net/http"
        //"net/url"
        "fmt"
	"io/ioutil"
	"github.com/swisskid/go-insteon/insteon"
	//"strings"
              "strconv"
)

func main() {
	insteon.PopulateAll()
        http.HandleFunc("/", foo)
        http.ListenAndServe(":9003", nil)
}
//Request Fields
type RequestMaster struct {
    Version			string		`json:"version"`
    Session			Session		`json:"session"`
    Request			Request		`json:"request"`
}
type Session struct{
    New				bool		`json:"new"`
    SessionId			string		`json:"sessionId"`
    Attributes			map[string]string	`json:"attributes"`
    Application			Application	`json:"application"`
    User			User		`json:"user"`
}

type Application struct {
    ApplicationId		string
}
type User struct{
    UserId			string
}
type Request struct { //I have to combine all the different requests into this. ugh.
    Type			string		`json:"type"` //LaunchRequest,IntentRequest,SessionEndedRequest
    Timestamp			string		`json:"date"`
    RequestId			string		`json:"requestId"`
    Intent			Intent		`json:"intent"`
    Reason			string		`json:"reason"`
}
type Intent struct {
    Name			string		`json:"name"`
    Slots			map[string]Slot		`json:"slots"`
}
type Slot struct {
    Name			string		`json:"name"`
    Value			string		`json:"value"`
}



//Response Fields
type ResponseMaster struct {
    Version			string		`json:"version"`
    SessionAttributes		map[string]interface{}	`json:"sessionAttributes,omitempty"`
    Response			Response	`json:"response"`
}

type Response struct {
    OutputSpeech		*OutputSpeech	`json:"outputSpeech,omitempty"`
    Card			*Card		`json:"card,omitempty"`
    Reprompt			*Reprompt	`json:"reprompt,omitempty"`
    ShouldEndSession		bool		`json:"shouldEndSession"`
}
type OutputSpeech struct {
    Type			string		`json:"type,omitempty"` //must omit empty so it doesn't print the whole object
    Text			string		`json:"text,omitempty"` //must omit empty so it doesn't print the whole object
}
type Card struct {
    Type			string		`json:"type,omitempty"` //must omit empty so it doesn't print the whole object
    Title			string		`json:"title,omitempty"`
    Content			string		`json:"content,omitempty"`
}
type Reprompt struct {
    OutputSpeech		OutputSpeech	`json:"outputSpeech,omitempty"`
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

	if l.Request.Intent.Name != "" {
	    fmt.Println("The name of the intent is " + l.Request.Intent.Name)
	    fmt.Println(l.Request.Intent.Slots)
	    for _, slot := range l.Request.Intent.Slots {
		fmt.Println(slot.Name)
		fmt.Println(slot.Value)
	    }
	}
	if l.Request.Intent.Name == "Lighting" { //Engage insteon subroutine
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
	    switch item_type{
		case "device":
		    n.Device_Id = id
		    n.Level = insteon.DevList[location].DimLevel/254 * 100
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
	    insteon.RunCommand(n)
	} else {

	    k.Text = "I am a working response"
	}
	if l.Request.Intent.Name == "Activate" { //Activate a scene
	    var n insteon.Command
	    n.Command = "on"
	    scene_name := l.Request.Intent.Slots["Scene"].Value
	    ret_type, scene_id, _ := insteon.SearchString(scene_name)
	    if ret_type == "scene" {
		n.Scene_Id = scene_id
		insteon.RunCommand(n)
		k.Text = "Turning the scene " + scene_name + " on"
	    } else {
		k.Text = "Failed to find scene " + scene_name 
	    }
	}
	j.Response.OutputSpeech = &k
	b, _ := json.Marshal(j)
	fmt.Println(string(b[:]))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("charset", "UTF-8")
	w.Write(b)
}
