package main

// necessary imports
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// using struct to handle json data
type User struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	DOB         string `json:"dob"`
	Address     string `json:"address"`
}

// this will initialize the databse
func initDB() (*sql.DB, error) {
	// please insert your own DB_USER_NAME and PASSWORD and DB_NAME
	connStr := "user=DB_USER_NAME password=DB_USER_PASSWORD dbname=DB_NAME sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	return db, nil
}

// this will create a new user
func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// validating the DOB
	_, err = time.Parse("2006-01-02", user.DOB)
	if err != nil {
		http.Error(w, "Invalid Date of Birth format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// connect database
	db, err := initDB()
	if err != nil {
		http.Error(w, fmt.Sprintf("Database connection failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// send the query
	query := `INSERT INTO users (first_name, last_name, email, phone_number, dob, address) 
			  VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	var userID int
	err = db.QueryRow(query, user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.DOB, user.Address).Scan(&userID)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		http.Error(w, fmt.Sprintf("Failed to insert user: %v", err), http.StatusInternalServerError)
		return
	}

	// showing back the response
	w.WriteHeader(http.StatusCreated)
	response := map[string]interface{}{
		"id":        userID,
		"full_name": user.FirstName + " " + user.LastName,
		"email":     user.Email,
	}
	json.NewEncoder(w).Encode(response)
}

// main function
func main() {
	// setting the router
	r := mux.NewRouter()

	r.HandleFunc("/api/users", createUser).Methods("POST")

	// starting the server
	fmt.Println("Server started at http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
