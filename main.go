package main

import (
	"os"
	"fmt"

	"net/http"
	"log"


	"html/template"

	"github.com/gorilla/websocket"

	"github.com/nats-io/go-nats"
	"time"
)


var upgrader = websocket.Upgrader{} // use default options

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };
    
    function t(){return (new Date()).toLocaleTimeString()};
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        console.log('connecting:',t())
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
        	console.log('connected:',t())
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
            console.log('receive:',t())

        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);

        console.log('send:',t())

        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))

var ws *websocket.Conn


func echo(w http.ResponseWriter, r *http.Request) {
	var err error
	ws, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer ws.Close()
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		
		
		log.Printf("ws recv: %s", message)
		
		nc.Publish("send",message)

	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func t()  {
	now := time.Now()
	year, mon, day := now.UTC().Date()
	hour, min, sec := now.UTC().Clock()
	zone, _ := now.UTC().Zone()
	fmt.Printf("UTC 时间是 %d-%d-%d %02d:%02d:%02d %s\n",
		year, mon, day, hour, min, sec, zone)
}


var nc *nats.Conn
// x nats 
func main()  {
	log.SetFlags(0)

	argsWithoutProg := os.Args[1:]

	fmt.Println(argsWithoutProg)


	nc, _ = nats.Connect(argsWithoutProg[0])



	nc.Subscribe("receive", func(m *nats.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
		//ws
		t()

		err := ws.WriteMessage(1, m.Data)
		if err != nil {
			log.Println("write:", err)
		}
	})

	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)


	log.Fatal(http.ListenAndServe(":8080", nil))

}
