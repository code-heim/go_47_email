package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
)

func sendEmail(to []string, subject string, body string) error {
	auth := smtp.PlainAuth(
		"",
		os.Getenv("FROM_EMAIL"),
		os.Getenv("FROM_EMAIL_PASSWORD"),
		os.Getenv("FROM_EMAIL_SMTP"),
	)

	message := "Subject: " + subject + "\n" + body
	return smtp.SendMail(
		os.Getenv("SMTP_ADDR"),
		auth,
		os.Getenv("FROM_EMAIL"),
		to,
		[]byte(message),
	)
}

func sendHtmlEmail(to []string, subject string, htmlBody string) error {
	auth := smtp.PlainAuth(
		"",
		os.Getenv("FROM_EMAIL"),
		os.Getenv("FROM_EMAIL_PASSWORD"),
		os.Getenv("FROM_EMAIL_SMTP"),
	)

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"

	message := "Subject: " + subject + "\n" + headers + "\n\n" + htmlBody
	return smtp.SendMail(
		os.Getenv("SMTP_ADDR"),
		auth,
		os.Getenv("FROM_EMAIL"),
		to,
		[]byte(message),
	)
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

// Define a struct to hold the incoming JSON request body
type EmailRequestBody struct {
	ToAddr  string `json:"to_addr"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func EmailHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON request body
	var reqBody EmailRequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Convert Param3 (comma-separated string) to a slice of strings
	to := strings.Split(reqBody.ToAddr, ",")

	err = sendEmail(to, reqBody.Subject, reqBody.Body)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully"))
}

func HTMLEmailHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON request body
	var reqBody EmailRequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Convert Param3 (comma-separated string) to a slice of strings
	to := strings.Split(reqBody.ToAddr, ",")

	err = sendHtmlEmail(to, reqBody.Subject, reqBody.Body)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully"))
}

type EmailWithTemplateRequestBody struct {
	ToAddr   string            `json:"to_addr"`
	Subject  string            `json:"subject"`
	Template string            `json:"template"`
	Vars     map[string]string `json:"vars"`
}

func HTMLTemplateEmailHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON request body
	var reqBody EmailWithTemplateRequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Convert Param3 (comma-separated string) to a slice of strings
	to := strings.Split(reqBody.ToAddr, ",")

	// Parse the HTML template
	tmpl, err := template.ParseFiles("./templates/" + reqBody.Template + ".html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	// Render the template with the map data
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, reqBody.Vars); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	log.Println(rendered.String())

	err = sendHtmlEmail(to, reqBody.Subject, rendered.String())
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully"))
}

func main() {
	godotenv.Load()
	addr := ":8080"

	mux := http.NewServeMux()
	mux.HandleFunc("/", HelloHandler)
	mux.HandleFunc("/email", EmailHandler)
	mux.HandleFunc("/html_email", HTMLEmailHandler)
	mux.HandleFunc("/html_email2", HTMLTemplateEmailHandler)

	log.Printf("server is listening at %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
