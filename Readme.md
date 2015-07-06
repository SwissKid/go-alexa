# Go Alexa
Quick go server to interact with Amazon Echo requests. 

Didn't deal with SSL, because I find it annoying. Proxying that bit through nginx to another port on my server. This handles and translates the Amazon Echo responses.

#### Insteon Integration
To make this work, you'll need my insteon library (and you'll need to update it with your API information for Insteon). You can get it with go get github.com/swisskid/go-insteon.

Alternatively, you can point to my webserver, https://veryoblivio.us/.

I'm including my server's public cert as well to make this easy.

#### Fuzzy integration
Alexa-fuzzy.go uses fuzzy matches based on the sound of phrases. You'll need to use:
 https://github.com/jeffsmith82/gofuzzy

#### Next Steps
I should probably list some future goals
* ~~Best case matching for insteon - Search the phrase, look for something "close enough" and if it's too far, ask for another prompt~~
* Actually follow amazon's rules for who can control it - checking date, source, certificate
* Doing SSL in the Go library - I bet this wouldn't even be that hard. I'm just lazy.
* ~~Doing tie-ins with insteon - Don't know how I'd do this yet, maybe give each person a key that they'd have to enter on my site to tie their username to a echo app? dunno.~~



