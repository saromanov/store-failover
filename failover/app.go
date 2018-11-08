package failover

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// App defines main struct for the program
type App struct {
	f    *Failover
	l    net.Listener
	m    sync.Mutex
	c    *Config
	quit chan struct{}
}

// New creates a new app
func New(c *Config) *App {
	f, err := NewRaft(c)
	if err != nil {
		panic(err)
	}
	lis, err := net.Listen("tcp", c.Addr)
	if err != nil {
		panic(err)
	}
	return &App{
		f:    f,
		c:    c,
		l:    lis,
		quit: make(chan struct{}),
	}
}

// Start provides init of the app
func (a *App) Start() {
	t := time.NewTicker(time.Duration(a.c.Interval) * time.Second)
	defer func() {
		t.Stop()
	}()

	go a.startHTTP()

	for {
		select {
		case <-t.C:
			a.checkCluster()
		case <-a.quit:
			return
		}
	}
}

// Close provides closing of the app
func (a *App) Close() {
	select {
	case <-a.quit:
		return
	default:
		break
	}

	close(a.quit)
}

// Get provides getting of the key
func (a *App) Get(key string) string {
	return a.f.Get(key)
}

// Set implements setting of the key-value pair to the raft store
func (a *App) Set(key, value string) error {
	return a.f.Set(key, value)
}

func (a *App) checkCluster() {
}

func (a *App) startHTTP() {
	if a.l == nil {
		return
	}

	m := mux.NewRouter()

	m.Handle("/master", &masterHandler{a})

	s := http.Server{
		Handler: m,
	}

	s.Serve(a.l)
}
