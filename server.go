package luxor

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq" // nolint
	"github.com/satori/go.uuid"
)

// TODO make these configurable
const (
	host     = "localhost"
	dbPort     = 5432
	dbname   = "luxor"
	user     = "luxor"
	password = "luxor"
)

// Server is mock Stratum V1 server
type Server struct {
	db                *sql.DB
	mutex             sync.Mutex
	extraNonceCounter int64
}

// Request is a Stratum request
type Request struct {
	ID     int
	Method string
	Params interface{}
}

// Start starts the server on the specified port. The callback
// ready is used to indicate the server is listening.
func (s *Server) Start(port int, ready func()) error {
	// TODO: move to database and read next val from there
	// this will get reset if the server is restarted
	s.extraNonceCounter = int64(1000000000)

	var err error
	s.db, err = sql.Open("postgres",
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, dbPort, user, password, dbname))
	if err != nil {
		return err
	}
	defer s.db.Close()

	if err := s.db.Ping(); err != nil {
		return err
	}

	// listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		return err
	}
	ready()
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	fmt.Printf("Serving %s\n", conn.RemoteAddr().String())

	for {
		netData, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println(err)
			continue
		}

		var req Request
		err = json.Unmarshal(([]byte(netData)), &req)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("req: %v\n", req)

		switch req.Method {
		case "mining.authorize":
			var authParams []interface{}
			authParams = req.Params.([]interface{})
			log.Printf("auth params: %v\n", authParams)
			username := authParams[0].(string)
			password := authParams[1].(string)

			passHash, err := hashPassword(password)
			if err != nil {
				log.Println(err)
				continue
			}

			query := `
INSERT INTO auth_requests (
  id, username, pass_hash
) VALUES (
  $1, $2, $3
)`
			if _, err := s.db.ExecContext(context.TODO(), query, req.ID, username, passHash); err != nil {
				log.Println(err)
				continue
			}

			conn.Write([]byte(fmt.Sprintf(`{"error": null, "id": %d, "result": true}`, req.ID) + "\n"))
		case "mining.subscribe":
			query := `
INSERT INTO subscriptions (
  id, subscription_id_1, subscription_id_2, extra_nonce_1
) VALUES (
  $1, $2, $3, $4
)`
			subscriptionID1 := strings.ReplaceAll(uuid.NewV4().String(), "-", "")
			subscriptionID2 := strings.ReplaceAll(uuid.NewV4().String(), "-", "")
			s.mutex.Lock()
			s.extraNonceCounter = s.extraNonceCounter + 1
			extraNonce1 := s.extraNonceCounter
			s.mutex.Unlock()
			encodedNonce1 := hex.EncodeToString([]byte(fmt.Sprintf("%d", extraNonce1)))
			if _, err := s.db.ExecContext(context.TODO(), query, req.ID, subscriptionID1, subscriptionID2, encodedNonce1); err != nil {
				log.Println(err)
				continue
			}

			conn.Write([]byte(fmt.Sprintf(`{"id": %d, "result": [[["mining.set_difficulty", "%s"], ["mining.notify", "%s"]], "%s", 4], "error": null}`,
				req.ID,
				subscriptionID1,
				subscriptionID2,
				encodedNonce1,
			) + "\n"))
		default:
			continue
		}

	}

	conn.Close()
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
