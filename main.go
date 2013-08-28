package main

import (
	"encoding/json"
	"fmt"
	"github.com/dynport/gocli"
	"github.com/dynport/gologger"
	"io/ioutil"
	"net/http"
	"os"
)

var logger = gologger.New()

func main() {
	if e := run(); e != nil {
		logger.Error(e.Error())
		os.Exit(1)
	}
}

type Server struct {
	Ip        string `json:"server_ip"`
	Number    int    `json:"server_number"`
	Name      string `json:"server_name"`
	Product   string `json:"product"`
	Dc        string `json:"dc"`
	Status    string `json:"status"`
	PaidUntil string `json:"paid_until"`
}

type Account struct {
	User, Password string
}

const (
	ENV_USER     = "HETZNER_USER"
	ENV_PASSWORD = "HETZNER_PASSWORD"
)

func AccountFromEnv() (account *Account, e error) {
	user := os.Getenv(ENV_USER)
	password := os.Getenv(ENV_PASSWORD)
	if user == "" || password == "" {
		return nil, fmt.Errorf("%s and %s must be set in env", ENV_USER, ENV_PASSWORD)
	}
	return &Account{User: user, Password: password}, nil
}

func init() {
	if os.Getenv("DEBUG") == "true" {
		logger.LogLevel = gologger.DEBUG
	}
}

func (account *Account) Servers() (servers []*Server, e error) {
	url := fmt.Sprintf("https://%s:%s@robot-ws.your-server.de/server", account.User, account.Password)
	logger.Debug("fetching servers")
	rsp, e := http.Get(url)
	if e != nil {
		return servers, e
	}
	defer rsp.Body.Close()
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return servers, e
	}
	if rsp.StatusCode != 200 {
		return nil, fmt.Errorf(string(b))
	}
	st := []map[string]*Server{}
	e = json.Unmarshal(b, &st)
	if e != nil {
		logger.Error(string(b))
		return servers, e
	}
	servers = []*Server{}
	for _, r := range st {
		if server, ok := r["server"]; ok {
			servers = append(servers, server)
		}
	}
	return servers, nil
}

func (server *Server) String() string {
	return fmt.Sprintf("%d: %s (%s)", server.Number, server.Name, server.Ip)
}

type ServerResponse struct {
	Servers []*Server `json:"server"`
}

func run() error {
	account, e := AccountFromEnv()
	if e != nil {
		return e
	}
	servers, e := account.Servers()
	if e != nil {
		return e
	}
	table := gocli.NewTable()
	for _, server := range servers {
		table.Add(server.Number, server.Name, server.Product, server.Ip, server.Status)
	}
	fmt.Println(table)
	return nil
}
