package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gocolly/colly"
)

// โครงสร้างสำหรับข้อมูลหนังสือที่เราจะส่งไปยัง API
type Book struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

func main() {
	// สร้าง Colly collector
	c := colly.NewCollector()

	// ดักจับข้อมูลจากหน้า Trending Books
	c.OnHTML("div.details", func(e *colly.HTMLElement) {
		// ดึงชื่อหนังสือ
		title := e.ChildText("h3.booktitle")
		// ดึงชื่อผู้เขียน
		authorName := e.ChildText("span.bookauthor a.results")

		// แสดงข้อมูลที่ดึงมา
		fmt.Printf("Book Title: %s, Author Name: %s \n", title, authorName)

		// สร้างโครงสร้างหนังสือ
		book := Book{
			Title:  title,
			Author: authorName,
		}

		// เรียกใช้ฟังก์ชันเพื่อส่งข้อมูลไปยัง API
		err := sendBookToAPI(book)
		if err != nil {
			log.Printf("Error sending book to API: %v", err)
		}
	})

	// เริ่มการดึงข้อมูลจาก URL ของหนังสือที่กำลังเป็นที่นิยม
	for i := 1; i <= 3; i++ {
		url := "https://openlibrary.org/trending/forever?page=" + strconv.Itoa(i)
		err := c.Visit(url)
		if err != nil {
			log.Fatal("Failed to visit URL:", err)
		}
	}
}

// ฟังก์ชันสำหรับส่งข้อมูลหนังสือไปยัง API
func sendBookToAPI(book Book) error {
	// แปลงข้อมูลหนังสือเป็น JSON
	bookJSON, err := json.Marshal(book)
	if err != nil {
		return fmt.Errorf("error marshalling book: %v", err)
	}

	// สร้างคำขอ POST เพื่อส่งข้อมูลไปยัง API
	resp, err := http.Post("http://localhost:8080/book", "application/json", bytes.NewBuffer(bookJSON))
	if err != nil {
		return fmt.Errorf("error making POST request: %v", err)
	}
	defer resp.Body.Close()

	// ตรวจสอบสถานะคำขอ
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API responded with status code: %d", resp.StatusCode)
	}

	fmt.Println("Book added successfully to the API!")
	return nil
}
