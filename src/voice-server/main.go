package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Nahemah1022/singsphere-voice-server/pkg/rtc"
	"github.com/Nahemah1022/singsphere-voice-server/pkg/socket"
	"github.com/Nahemah1022/singsphere-voice-server/room"
	"github.com/Nahemah1022/singsphere-voice-server/user"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var roomManager *room.RoomManager

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	roomManager = room.NewRoomManager()
	router := registerRouters()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
		log.Print("Port unset, use default port 8000")
	}
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("listening on %s\n", addr)

	server := &http.Server{
		Handler:      router,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func registerRouters() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/stats", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		bytes, err := json.Marshal(roomManager.GetStats())
		if err != nil {
			http.Error(w, fmt.Sprint(err), 500)
		}
		w.Write(bytes)
	}).Methods("GET")

	router.HandleFunc("/api/rooms/{id}", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		vars := mux.Vars(req)
		roomID := vars["id"]

		r, err := roomManager.Get(roomID)
		if err == room.ErrNotFound {
			http.NotFound(w, req)
			return
		}

		bytes, err := json.Marshal(r.Wrap())
		if err != nil {
			http.Error(w, fmt.Sprint(err), 500)
		}
		w.Write(bytes)
	}).Methods("GET")

	router.HandleFunc("/ws/{id}", func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		roomID := vars["id"]
		room := roomManager.GetOrCreate(roomID)

		// Establish websocket connection and inject it as external dependency to user
		ws, err := socket.New(w, req, func() {
			log.Println("ws closed")
		})
		if err != nil {
			log.Println(err)
			return
		}

		// Establish WebRTC peer conection and inject it as external dependency to user
		rtcNode, err := rtc.New()
		if err != nil {
			log.Println(err)
			return
		}

		newUser := user.New(room.UserJoinCh, room.UserLeaveCh, ws, rtcNode)
		go newUser.Run()
	})

	return router
}
