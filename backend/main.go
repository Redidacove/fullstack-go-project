package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type User struct {
	Id  int	`json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

//main function
func main(){
	db, err :=sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("api/go/users", getUsers(db)).Methods("GET")
	router.HandleFunc("api/go/users", createUsers(db)).Methods("POST")
	router.HandleFunc("api/go/users{id}", getUser(db)).Methods("GET")
	router.HandleFunc("api/go/users{id}", updateUser(db)).Methods("PUT")
	router.HandleFunc("api/go/users{id}", deleteUser(db)).Methods("DELETE")

	//wrap router with CORS and JSON
	enhancedRouter := enableCORS(jsonContentTypeMiddleware(router))

	//start server
	log.Fatal(http.ListenAndServe(":8000", enhancedRouter))
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		// Set CORS Header
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Check if this is a preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Pass down the request to the next middleware(or final handler)
		next.ServeHTTP(w, r)
	})
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		// Set JSON Content-Type 
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		rows, err := db.Query("SELECT * FROM users")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		users := []User{} // array of users
		for rows.Next(){
			var user User
			if err := rows.Scan(&user.Id, &user.Name, &user.Email); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			users = append(users, user)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(users); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		vars := mux.Vars(r)
		id := vars["id"]

		var user User
		err := db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&user.Id, &user.Name, &user.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func createUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err := db.QueryRow("INSERT INTO users(name, email) VALUES($1, $2) RETURNING id", user.Name, user.Email).Scan(&user.Id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		vars := mux.Vars(r)
		id := vars["id"]

		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := db.Exec("UPDATE users SET name = $1, email = $2 WHERE id = $3", user.Name, user.Email, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Retrive the updated User after updation
		var updatedUser User
		err = db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&updatedUser.Id, &updatedUser.Name, &updatedUser.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Send the updated User as response
		json.NewEncoder(w).Encode(updatedUser)
	}
}

func deleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		vars := mux.Vars(r)
		id := vars["id"]

		var user User
		err := db.QueryRow("SELECT * FROM USERS WHERE id = $1",id).Scan(&user.Id,&user.Name,&user.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
			}else {
				_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				json.NewEncoder(w).Encode("User Deleted")
		}
	}
}
