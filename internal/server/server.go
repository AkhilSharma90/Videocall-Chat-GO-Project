package server

import (
	"flag"
	"os"
	"time"

	"v/internal/handlers"
	w "v/pkg/webrtc"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"
	"github.com/gofiber/websocket/v2"
)

var (
	addr = flag.String("addr", ":"+os.Getenv("PORT"), "") // Address to listen on, defaults to the value of the PORT environment variable
	cert = flag.String("cert", "", "")  // TLS certificate file path
	key  = flag.String("key", "", "")    // TLS private key file path
)

func Run() error {
	flag.Parse()

	if *addr == ":" {
		*addr = ":8080" // Set default address if not provided
	}

	engine := html.New("./views", ".html") // Create HTML template engine
	app := fiber.New(fiber.Config{Views: engine})
	app.Use(logger.New())  // Enable request logging
	app.Use(cors.New())    // Enable CORS

	// Define routes and handlers
	app.Get("/", handlers.Welcome)
	app.Get("/room/create", handlers.RoomCreate)
	app.Get("/room/:uuid", handlers.Room)
	app.Get("/room/:uuid/websocket", websocket.New(handlers.RoomWebsocket, websocket.Config{
		HandshakeTimeout: 10 * time.Second,
	}))
	app.Get("/room/:uuid/chat", handlers.RoomChat)
	app.Get("/room/:uuid/chat/websocket", websocket.New(handlers.RoomChatWebsocket))
	app.Get("/room/:uuid/viewer/websocket", websocket.New(handlers.RoomViewerWebsocket))
	app.Get("/stream/:suuid", handlers.Stream)
	app.Get("/stream/:suuid/websocket", websocket.New(handlers.StreamWebsocket, websocket.Config{
		HandshakeTimeout: 10 * time.Second,
	}))
	app.Get("/stream/:suuid/chat/websocket", websocket.New(handlers.StreamChatWebsocket))
	app.Get("/stream/:suuid/viewer/websocket", websocket.New(handlers.StreamViewerWebsocket))
	app.Static("/", "./assets")  // Serve static assets

	w.Rooms = make(map[string]*w.Room)
	w.Streams = make(map[string]*w.Room)
	go dispatchKeyFrames()   // Start a goroutine to dispatch key frames periodically
	if *cert != "" {
		return app.ListenTLS(*addr, *cert, *key) // Start the server with TLS if certificate and key are provided
	}
	return app.Listen(*addr) // Start the server without TLS
}

func dispatchKeyFrames() {
	for range time.NewTicker(time.Second * 3).C {
		for _, room := range w.Rooms {
			room.Peers.DispatchKeyFrame() // Dispatch key frames to all peers in each room
		}
	}
}
