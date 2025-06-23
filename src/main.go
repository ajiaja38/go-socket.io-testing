package main

import (
	"net/http"
	"os"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type ResponseEntity struct {
	Message  string    `json:"message"`
	DateTime time.Time `json:"date_time"`
}

var log *logrus.Logger = logrus.New()

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(c socketio.Conn) error {
		log.Infof("✅ connected to socket %s", c.ID())
		return nil
	})

	server.OnDisconnect("/", func(c socketio.Conn, s string) {
		log.Infof("❌ disconnected: %s %s", c.ID(), s)
	})

	c := cron.New()

	c.AddFunc("@every 10s", func() {
		payload := ResponseEntity{
			Message:  "Halo dari golang socket io ✅",
			DateTime: time.Now(),
		}

		server.BroadcastToNamespace("/", "periodic-message", payload)
		log.Infof("✅ Success Send Payload to Socket.io")
	})

	c.Start()

	go server.Serve()
	defer server.Close()

	appPort := os.Getenv("PORT")
	if appPort == "" {
		appPort = "8080"
		log.Warnf("PORT environment variable not set, defaulting to %s", appPort)
	}
	port := ":" + appPort

	mux := http.NewServeMux()
	mux.Handle("/socket.io/", server)
	mux.Handle("/", http.FileServer(http.Dir("./public")))

	log.Infof("🚀 Application listening on http://localhost%s", port)

	err := http.ListenAndServe(port, mux)

	if err != nil {
		log.Fatalf("Fatal: HTTP Server failed to start: %v", err)
	}
}
