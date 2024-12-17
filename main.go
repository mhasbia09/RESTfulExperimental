package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"RESTfulExperimental/models"

	_ "github.com/go-sql-driver/mysql"
)

var (
    db  *sql.DB
    mu  sync.Mutex
)

func init() {
    var err error
    // Setup koneksi ke database MySQL
    dsn := "root:@tcp(127.0.0.1:3306)/restfulexperiment" // Ganti username dan password sesuai dengan konfigurasi Anda
    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }

    // Cek koneksi
    if err = db.Ping(); err != nil {
        log.Fatal(err)
    }
}

func getItems(w http.ResponseWriter, r *http.Request) {
    mu.Lock()
    defer mu.Unlock()

    rows, err := db.Query("SELECT id, name, price FROM items")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var itemList []models.Item
    for rows.Next() {
        var item models.Item
        if err := rows.Scan(&item.ID, &item.Name, &item.Price); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        itemList = append(itemList, item)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(itemList)
}

func createItem(w http.ResponseWriter, r *http.Request) {
    var item models.Item
    json.NewDecoder(r.Body).Decode(&item)

    mu.Lock()
    defer mu.Unlock()

    _, err := db.Exec("INSERT INTO items (id, name, price) VALUES (?, ?, ?)", item.ID, item.Name, item.Price)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(item)
}

func updateItem(w http.ResponseWriter, r *http.Request) {
    var item models.Item
    json.NewDecoder(r.Body).Decode(&item)

    mu.Lock()
    defer mu.Unlock()

    _, err := db.Exec("UPDATE items SET name = ?, price = ? WHERE id = ?", item.Name, item.Price, item.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(item)
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")

    mu.Lock()
    defer mu.Unlock()

    _, err := db.Exec("DELETE FROM items WHERE id = ?", id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func main() {
    defer db.Close() // Pastikan untuk menutup koneksi saat aplikasi selesai

    http.HandleFunc("/items", getItems)
    http.HandleFunc("/items/create", createItem)
    http.HandleFunc("/items/update", updateItem)
    http.HandleFunc("/items/delete", deleteItem)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    log.Println("Server is running on :8080")
    http.ListenAndServe(":8080", nil)
}
