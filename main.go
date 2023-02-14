package main

import (
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			o := r.Header.Get("Origin")
			return o == *origin //"http://127.0.0.1:8080"
		},
	}
	channels  = &Channels{}
	addrWs    = flag.String("wsaddr", ":8080", "websocket service address")
	addrHt    = flag.String("addr", ":8000", "http service address")
	origin    = flag.String("origin", "http://localhost:8080", "origin for check")
	wsendpoint = flag.String("endpoint", "/ws", "websocket endpoint")
	homeTempl = template.Must(template.New("").Parse(homeHTML))
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	chname := r.FormValue("channel")
	wrap := NewWS(ws)
	channels.Add(chname, wrap)
	wrap.ListenAndServe()
}

func wsRoot(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	m := make(map[string]string)
	m["Channel"] = r.FormValue("channel")
	if m["Channel"] == "" {
		http.Redirect(w, r, "/?channel=test", http.StatusFound)
		return
	}
	h, p, _ := net.SplitHostPort(*addrWs)
	if h == "" {
		h = "localhost"
	}
	m["Host"] = net.JoinHostPort(h, p)
	homeTempl.Execute(w, m)
}

func receiverMsg(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	chname := r.FormValue("channel")
	msg := []byte(r.FormValue("msg"))
	channels.Send(chname, msg)
	w.WriteHeader(http.StatusOK)
}

func main() {
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("/", wsRoot)
	mux.HandleFunc(*wsendpoint, wsHandler)
	go http.ListenAndServe(*addrWs, mux)
	http.ListenAndServe(*addrHt, http.HandlerFunc(receiverMsg))
}

const homeHTML = `<!DOCTYPE html>
<html lang="en">
    <head>
        <title>WebSocket Example</title>
    </head>
    <body>
        <pre id="test">&nbsp;</pre>
        <script type="text/javascript">
            (function() {
                var data = document.getElementById("test");
                var conn = new WebSocket("ws://{{.Host}}/ws?channel={{.Channel}}");
                conn.onclose = function(evt) {
                    data.textContent = 'Connection closed';
                }
                conn.onmessage = function(evt) {
                    console.log('msg received');
                    data.textContent = evt.data;
                }
            })();
        </script>
    </body>
</html>
`
