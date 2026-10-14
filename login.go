package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // Driver para MariaDB/MySQL
)

// Estructura para recibir el JSON de entrada
type LoginRequest struct {
	Usuario string `json:"usuario"`
	Clave   string `json:"clave"`
	Puesto  string `json:"puesto"`
}

// Estructura para devolver la información relevante
type UserResponse struct {
	Estatura       float64 `json:"estatura"`
	Peso           float64 `json:"peso"`
	HorarioLaboral string  `json:"horario_laboral"`
	Edad           int     `json:"edad"`
	FechaInicio    string  `json:"fecha_inicio"`
}

var db *sql.DB

func main() {
	var err error

	// Configura tu cadena de conexión: usuario:contraseña@tcp(host:puerto)/nombre_bd
	// parseTime=true es importante para manejar correctamente los campos DATE/DATETIME
	dsn := "root:123456@tcp(127.0.0.1:3306)/gestiondb?parseTime=true"

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error al abrir la conexión con la base de datos: %v", err)
	}
	defer db.Close()

	// Validar la conexión
	if err = db.Ping(); err != nil {
		log.Fatalf("No se pudo conectar a MariaDB: %v", err)
	}
	log.Println("Conexión exitosa a MariaDB")

	// Enrutador HTTP (compatible con enrutamiento nativo de Go 1.22+)
	http.HandleFunc("POST /api/login", handleLogin)

	log.Println("Servidor HTTP corriendo en el puerto 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Decodificar el JSON recibido en la solicitud
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Formato JSON inválido"})
		return
	}

	// Validar campos vacíos básicos
	if req.Usuario == "" || req.Clave == "" || req.Puesto == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Usuario, clave y puesto son obligatorios"})
		return
	}

	// 2. Consultar la base de datos para validar credenciales y puesto
	// Nota de seguridad: En producción, las claves deben estar hasheadas (ej. con bcrypt)
	query := `
		SELECT estatura, peso, horario_laboral, edad, DATE_FORMAT(fecha_inicio, '%Y-%m-%d') 
		FROM empleados 
		WHERE usuario = ? AND clave = ? AND puesto = ?
	`

	var resp UserResponse
	err := db.QueryRow(query, req.Usuario, req.Clave, req.Puesto).Scan(
		&resp.Estatura,
		&resp.Peso,
		&resp.HorarioLaboral,
		&resp.Edad,
		&resp.FechaInicio,
	)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Credenciales o puesto incorrectos"})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error interno en el servidor"})
		log.Println("Error en consulta SQL:", err)
		return
	}

	// 3. Devolver la información solicitada en formato JSON
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
