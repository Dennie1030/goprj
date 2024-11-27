package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// 创建用户表
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	// 创建数据表
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS user_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		data_name TEXT NOT NULL,
		data_value TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	)`)
	if err != nil {
		log.Fatalf("Failed to create user_data table: %v", err)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// 插入用户数据
	_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, hashedPassword)
	if err != nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "User registered successfully")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", user.Username).Scan(&hashedPassword)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Login successful")
}

func uploadDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求数据
	var requestData struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		DataName  string `json:"dataName"`
		DataValue string `json:"dataValue"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证用户
	var userID int
	var hashedPassword string
	err := db.QueryRow("SELECT id, password FROM users WHERE username = ?", requestData.Username).Scan(&userID, &hashedPassword)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(requestData.Password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 检查是否存在相同的 dataName
	var existingID int
	err = db.QueryRow("SELECT id FROM user_data WHERE user_id = ? AND data_name = ?", userID, requestData.DataName).Scan(&existingID)
	if err == nil {
		// 如果找到数据，执行更新
		_, err = db.Exec("UPDATE user_data SET data_value = ? WHERE id = ?", requestData.DataValue, existingID)
		if err != nil {
			http.Error(w, "Failed to update data", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Data updated successfully")
		return
	} else if err != sql.ErrNoRows {
		// 处理其他查询错误
		http.Error(w, "Failed to check data existence", http.StatusInternalServerError)
		return
	}

	// 如果不存在，则插入数据
	_, err = db.Exec("INSERT INTO user_data (user_id, data_name, data_value) VALUES (?, ?, ?)",
		userID, requestData.DataName, requestData.DataValue)
	if err != nil {
		http.Error(w, "Failed to upload data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Data uploaded successfully")
}

func getDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 获取查询参数
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	dataName := r.URL.Query().Get("dataName")

	if username == "" || password == "" || dataName == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// 验证用户
	var userID int
	var hashedPassword string
	err := db.QueryRow("SELECT id, password FROM users WHERE username = ?", username).Scan(&userID, &hashedPassword)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 查询对应的 dataValue
	var dataValue string
	err = db.QueryRow("SELECT data_value FROM user_data WHERE user_id = ? AND data_name = ?", userID, dataName).Scan(&dataValue)
	if err == sql.ErrNoRows {
		http.Error(w, "Data not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	// 返回数据
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"dataName":  dataName,
		"dataValue": dataValue,
	})
}

func deleteDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	dataName := r.URL.Query().Get("dataname")
	if username == "" || dataName == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	_, err = db.Exec("DELETE FROM user_data WHERE user_id = ? AND data_name = ?", userID, dataName)
	if err != nil {
		http.Error(w, "Failed to delete data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Data deleted successfully")
}

func getDataNamesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 获取用户名参数
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing 'username' parameter", http.StatusBadRequest)
		return
	}

	// 查询用户 ID
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 查询用户的数据名称列表
	rows, err := db.Query("SELECT data_name FROM user_data WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to fetch data names", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 收集数据名称
	var dataNames []string
	for rows.Next() {
		var dataName string
		if err := rows.Scan(&dataName); err != nil {
			http.Error(w, "Failed to parse data names", http.StatusInternalServerError)
			return
		}
		dataNames = append(dataNames, dataName)
	}

	// 返回数据名称列表
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dataNames)
}

func main() {

	initDB()
	// 路由处理
	http.HandleFunc("/upload", uploadDataHandler)
	http.HandleFunc("/datanames", getDataNamesHandler)
	http.HandleFunc("/data", getDataHandler)
	http.HandleFunc("/delete", deleteDataHandler)
	http.HandleFunc("/register", registerHandler)

	fmt.Println("Server is running on :9080")
	if err := http.ListenAndServe(":9080", nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
