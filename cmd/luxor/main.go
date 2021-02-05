package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"golang.org/x/crypto/bcrypt"

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

			passHash, err := hashPassword(password)
			if err != nil {
				log.Println(err)
				return
			}

			if _, err := db.ExecContext(context.TODO(), "INSERT INTO auth_requests VALUES ($1, $2, $3, NOW())", req.ID, username, passHash); err != nil {
				log.Println(err)
				return
			}
		default:
		}

		conn.Write([]byte(fmt.Sprintf(`{"error": null, "id": %d, "result": true}`, req.ID)))
	}
	conn.Close()
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
