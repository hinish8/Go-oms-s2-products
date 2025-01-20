package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	DBHost  = "127.0.0.1"
	DBPort  = 5432
	DBUser  = "root"
	DBPass  = "p@ssword"
	DBName  = "service2"
	AppPort = ":8080"
)

var db *sql.DB

type Customer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func connectToDatabase() error {
	pgConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		DBHost, DBPort, DBUser, DBPass, DBName)

	var err error

	db, err = sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v\n", err)
		return err
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Unable to establish connection to PostgreSQL: %v\n", err)
		return err
	}

	fmt.Println("Successfully connected to the database.")
	return nil
}

func getCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID := vars["id"]

	var c Customer
	err := db.QueryRow("SELECT id, name FROM customer WHERE id = $1", customerID).Scan(&c.ID, &c.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch customer", http.StatusInternalServerError)
		log.Printf("Error fetching customer: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]
	fmt.Println(productID)
	var p Product
	err := db.QueryRow("SELECT id, name FROM product WHERE id = $1", productID).Scan(&p.ID, &p.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch product", http.StatusInternalServerError)
		log.Printf("Error fetching product: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	var p Product
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO product (id, name) VALUES ($1, $2)", p.ID, p.Name)
	if err != nil {
		http.Error(w, "Failed to add product", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Product created succesfully %d", result)
}

func createCustomer(w http.ResponseWriter, r *http.Request) {
	var c Customer
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO customer (id, name) VALUES ($1, $2)", c.ID, c.Name)
	if err != nil {
		http.Error(w, "Failed to add customer", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Customer created succesfully %d", result)
}

func getAllProduct(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("SELECT id, name FROM product")
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch product", http.StatusInternalServerError)
		log.Printf("Error fetching product: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Name)
		if err != nil {
			log.Fatalf("Error scanning order: %v\n", err)
		}
		jsonData, err := json.Marshal(p)
		if err != nil {
			http.Error(w, "Failed to serialize product data", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func getAllCustomer(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("SELECT id, name FROM customer")
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch product", http.StatusInternalServerError)
		log.Printf("Error fetching product: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var c Customer
		err := rows.Scan(&c.ID, &c.Name)
		if err != nil {
			log.Fatalf("Error scanning order: %v\n", err)
		}
		jsonData, err := json.Marshal(c)
		if err != nil {
			http.Error(w, "Failed to serialize customer data", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func main() {
	if err := connectToDatabase(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/customer/{id}", getCustomer).Methods("GET")
	r.HandleFunc("/product/{id}", getProduct).Methods("GET")
	r.HandleFunc("/All_product", getAllProduct).Methods("GET")
	r.HandleFunc("/All_customer", getAllCustomer).Methods("GET")
	r.HandleFunc("/create_product", createProduct).Methods("POST")
	r.HandleFunc("/create_customer", createCustomer).Methods("POST")

	fmt.Println("Service 2 running on port 8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}
