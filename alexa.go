package main

import (
        //      "io"
        "encoding/json"
        "net/http"
        //"net/url"
        "fmt"
	"io/ioutil"
        //      "strconv"
)

func main() {
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
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Println(string(body[:]))
	var l RequestMaster
	json.Unmarshal(body, &l)
	if l.Request.Intent.Name != "" {
	    fmt.Println("The name of the intent is " + l.Request.Intent.Name)
	    fmt.Println(l.Request.Intent.Slots)
	    for _, slot := range l.Request.Intent.Slots {
		fmt.Println(slot.Name)
		fmt.Println(slot.Value)
	    }
	}
	fmt.Println(l)
	var j ResponseMaster
	var k OutputSpeech
	j.Version = "1.0"
	k.Type = "PlainText"
	k.Text = "I am a working response"
	j.Response.OutputSpeech = &k
	j.Response.ShouldEndSession = true
	b, _ := json.Marshal(j)
	fmt.Println(string(b[:]))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("charset", "UTF-8")
	w.Write(b)
}
