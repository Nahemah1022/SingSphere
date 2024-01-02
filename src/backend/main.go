package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Nahemah1022/singsphere-backend/user"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	fmt.Println(os.Getenv("PORT"))
	rooms := user.NewRooms()
	router := mux.NewRouter()

	router.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		bytes, err := json.Marshal(rooms.GetStats())
		if err != nil {
			http.Error(w, fmt.Sprint(err), 500)
		}
		w.Write(bytes)
	}).Methods("GET")

	router.HandleFunc("/api/rooms/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		vars := mux.Vars(r)
		roomID := vars["id"]
		room, err := rooms.Get(roomID)
		if err == user.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		bytes, err := json.Marshal(room.Wrap(nil))
		if err != nil {
			http.Error(w, fmt.Sprint(err), 500)
		}
		w.Write(bytes)
	}).Methods("GET")

	router.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		user.ServeWs(rooms, w, r)
	})

	// go rooms.Watch()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
		log.Printf("Defaulting to port %s", port)
	}
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("listening on %s\n", addr)

	srv := &http.Server{
		Handler:      router,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
