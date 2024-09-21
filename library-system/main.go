package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
)

var db *sql.DB

func init() {
    // โหลดค่าในไฟล์ .env
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }
}

// Struct สำหรับหนังสือ
type Book struct {
    ID        int    `json:"id"`
    Title     string `json:"title"`
    Author    string `json:"author"`
    Available bool   `json:"available"`
}

// Struct สำหรับผู้ใช้
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Struct สำหรับการยืมหนังสือ
type Borrow struct {
    ID         int       `json:"id"`
    UserID     int       `json:"user_id"`
    BookID     int       `json:"book_id"`
    BorrowDate time.Time `json:"borrow_date"`
    ReturnDate time.Time `json:"return_date"`
}

func main() {
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbName := os.Getenv("DB_NAME")

    // สร้าง connection string
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)

    var err error
    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    router := mux.NewRouter()

    router.HandleFunc("/books", getBooks).Methods("GET")
    router.HandleFunc("/book", addBook).Methods("POST")
    router.HandleFunc("/users", getUsers).Methods("GET")
    router.HandleFunc("/register", addUser).Methods("POST")
    router.HandleFunc("/borrow", borrowBook).Methods("POST")
    router.HandleFunc("/return", returnBook).Methods("POST")

    log.Fatal(http.ListenAndServe(":8080", router))
}

// ฟังก์ชันดึงข้อมูลหนังสือ
func getBooks(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, title, author, available FROM books")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var books []Book
    for rows.Next() {
        var book Book
        err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Available)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        books = append(books, book)
    }

    json.NewEncoder(w).Encode(books)
}

// ฟังก์ชันเพิ่มหนังสือ
func addBook(w http.ResponseWriter, r *http.Request) {
    var book Book
    json.NewDecoder(r.Body).Decode(&book)

    query := "INSERT INTO books (title, author) VALUES (?, ?)"
    _, err := db.Exec(query, book.Title, book.Author)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintln(w, "Book added successfully")
}

// ฟังก์ชันดึงข้อมูลผู้ใช้
func getUsers(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, name, email FROM users")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Name, &user.Email)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        users = append(users, user)
    }

    json.NewEncoder(w).Encode(users)
}

// ฟังก์ชันเพิ่มผู้ใช้
func addUser(w http.ResponseWriter, r *http.Request) {
    var user User
    json.NewDecoder(r.Body).Decode(&user)

    query := "INSERT INTO users (name, email) VALUES (?, ?)"
    _, err := db.Exec(query, user.Name, user.Email)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintln(w, "User added successfully")
}

// ฟังก์ชันยืมหนังสือ
func borrowBook(w http.ResponseWriter, r *http.Request) {
    var borrow Borrow
    json.NewDecoder(r.Body).Decode(&borrow)

    query := "INSERT INTO borrows (user_id, book_id, borrow_date, return_date) VALUES (?, ?, CURDATE(), NULL)"
    _, err := db.Exec(query, borrow.UserID, borrow.BookID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    query = "UPDATE books SET available = FALSE WHERE id = ?"
    _, err = db.Exec(query, borrow.BookID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintln(w, "Book borrowed successfully")
}

// ฟังก์ชันคืนหนังสือ
func returnBook(w http.ResponseWriter, r *http.Request) {
    var borrow Borrow
    json.NewDecoder(r.Body).Decode(&borrow)

    query := "UPDATE borrows SET return_date = CURDATE() WHERE user_id = ? AND book_id = ? AND return_date IS NULL"
    _, err := db.Exec(query, borrow.UserID, borrow.BookID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    query = "UPDATE books SET available = TRUE WHERE id = ?"
    _, err = db.Exec(query, borrow.BookID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintln(w, "Book returned successfully")
}
