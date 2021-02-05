package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	dbname   = "luxor"
	user     = "luxor"
	password = "luxor"
)

type AuthorizeParams struct {
	Username, Password int
}

type Mining struct{}

func (m *Mining) Authorize(params *AuthorizeParams, reply *bool) error {
	*reply = true
	return nil
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres",
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handle(conn)
	}
}

type Request struct {
	ID     int
	Method string
	Params interface{}
}

func handle(conn net.Conn) {
	fmt.Printf("Serving %s\n", conn.RemoteAddr().String())
	for {
		netData, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		temp := strings.TrimSpace(string(netData))

		var req Request
		err = json.Unmarshal(([]byte(netData)), &req)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("%v\n", req)

		switch req.Method {
		case "mining.authorize":
			var authParams []interface{}
			authParams = req.Params.([]interface{})
			fmt.Printf("%v\n", authParams)
			username := authParams[0].(string)
			password := authParams[1].(string)
			fmt.Println(username, password)
			if _, err := db.QueryContext(context.TODO(), "SELECT * FROM auth_requests"); err != nil {
				log.Println(err)
				return
			}

			if _, err := db.ExecContext(context.TODO(), "INSERT INTO auth_requests VALUES ($1, $2, $3, NOW())", req.ID, username, password); err != nil {
				log.Println(err)
				return
			}
		default:
		}

		result := temp + "\n"
		conn.Write([]byte(string(result)))
	}
	conn.Close()
}

// func main() {
// 	m := new(Mining)
// 	server := rpc.NewServer()
// 	server.RegisterName("mining", m)
// 	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
// 	listener, e := net.Listen("tcp", ":1234")
// 	if e != nil {
// 		log.Fatal("listen error:", e)
// 	}
// 	for {
// 		if conn, err := listener.Accept(); err != nil {
// 			log.Fatal("accept error: " + err.Error())
// 		} else {
// 			log.Printf("new connection established\n")
// 			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
// 		}
// 	}
// }

// func main() {
// 	http.HandleFunc("/", handler)
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

// func handler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
// 	bytes, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	fmt.Println(string(bytes))
// }
