package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// Patient structure
type Patient struct {
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	BirthDate    string `json:"birthDate"`
	Gender       string `json:"gender"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
}

// AWS SES email sender
func sendEmail(patient Patient) error {
	sesServiceURL := os.Getenv("AWS_SES_URL")
	emailData := map[string]string{
		"to":      patient.Email,
		"subject": "Bienvenido a MediSync",
		"body":    fmt.Sprintf("Hola %s, bienvenido a MediSync! Tu usuario es: %s y tu contraseña es: %s", patient.FirstName, patient.Username, patient.PasswordHash),
	}

	jsonData, _ := json.Marshal(emailData)
	resp, err := http.Post(sesServiceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error sending email, status: %d", resp.StatusCode)
	}

	return nil
}

// Register patient handler
func registerPatientHandler(w http.ResponseWriter, r *http.Request) {
	var patient Patient
	if err := json.NewDecoder(r.Body).Decode(&patient); err != nil {
		http.Error(w, "invalid input data", http.StatusBadRequest)
		return
	}

	patientServiceURL := os.Getenv("PATIENT_SERVICE_URL") + "/create-patient/patients"
	jsonData, _ := json.Marshal(patient)
	resp, err := http.Post(patientServiceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "error registering patient", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Send email notification
	if err := sendEmail(patient); err != nil {
		log.Printf("Error sending email: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "patient registered successfully"}`))
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/register", registerPatientHandler).Methods("POST")
	http.HandleFunc("/healthz", HealthCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
