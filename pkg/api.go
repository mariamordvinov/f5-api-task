package api_sec

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("my_secret_key")

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

// make sure that role is user or admin only
func isRoleValid(role string) bool {
	if role == "admin" || role == "user" {
		return true
	}
	return false
}

// return true if username already exists in users
func isUsernameExists(username string) bool {
	for _, user := range users {
		if username == user.Username {
			return true
		}
	}
	return false
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// make sure that role is user or admin only (input validation)
	if !isRoleValid(user.Role) {
		http.Error(w, "Invalid role !", http.StatusBadRequest)
		return
	}

	//Made username unique
	if isUsernameExists(user.Username) {
		http.Error(w, "username already exists", http.StatusBadRequest)
		return
	}

	// Defining IDs this way may cause security issues. it makes it easy for someone to guess resources IDs.
	// I would define user id as a UUID so it wont be easy to guess
	//
	// I didnt implement the change to UUID because in the api_usage_example i recived, the response example contains the id as int, and I wanted to follow it exactly.
	// But i would do something like this:
	// user.ID = uuid.New()

	user.ID = len(users) + 1

	//saving raw passwords is bad security wise.
	// I would change the way i store the password to storing a hash of the password.
	// I also didnt implement it because in the api_usage_example i recived the response example contains the raw password, and I wanted to follow it exactly.

	users = append(users, user)
	json.NewEncoder(w).Encode(user)
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var creds User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Authenticate user
	var authenticatedUser *User
	for _, user := range users {
		//if the password was stored as hash, like I suggested, then to check if the password matches I will compare hases.
		if user.Username == creds.Username && user.Password == creds.Password {
			authenticatedUser = &user
			break
		}
	}
	if authenticatedUser == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		Username: authenticatedUser.Username,
		Role:     authenticatedUser.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func GetUseres(w http.ResponseWriter, r *http.Request, claims *Claims) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

	//only admins are allowd to get users
	if claims.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func AccountsHandler(w http.ResponseWriter, r *http.Request, claims *Claims) {

	// both create account and list accounts are allowed only to admins. So I moved the condition to cover both cases
	if claims.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		createAccount(w, r, claims)
		return
	}
	if r.Method == http.MethodGet {
		listAccounts(w, r, claims)
		return
	}

	//in case got unsupported method inform the client
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func createAccount(w http.ResponseWriter, r *http.Request, claims *Claims) {
	var acc Account
	if err := json.NewDecoder(r.Body).Decode(&acc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Defining IDs this way may cause security issues. like I mentioned before I would use UUID.
	acc.ID = len(accounts) + 1
	acc.CreatedAt = time.Now()
	accounts = append(accounts, acc)
	json.NewEncoder(w).Encode(acc)
}

func listAccounts(w http.ResponseWriter, r *http.Request, claims *Claims) {
	json.NewEncoder(w).Encode(accounts)
}

func BalanceHandler(w http.ResponseWriter, r *http.Request, claims *Claims) {

	// only users are allowed to deposit and withdraw, so added enforcment.
	if claims.Role != "user" && r.Method != http.MethodGet {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}
	switch r.Method {
	case http.MethodGet:
		getBalance(w, r, claims)
	case http.MethodPost:
		depositBalance(w, r, claims)
	case http.MethodDelete:
		withdrawBalance(w, r, claims)
	}
}

// checks if given user ID matches the user according to the claims
func isUserIdValid(claims *Claims, userId int) bool {
	for _, user := range users {
		if claims.Username == user.Username && user.ID == userId {
			return true
		}
	}
	return false
}

func getBalance(w http.ResponseWriter, r *http.Request, claims *Claims) {
	userId := r.URL.Query().Get("user_id")
	uid, _ := strconv.Atoi(userId)
	// added this check to prevent BOLA- without this validation a user can access data that belongs to another user (unauthorized access)

	if !isUserIdValid(claims, uid) && claims.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	for _, acc := range accounts {
		if acc.UserID == uid {
			json.NewEncoder(w).Encode(map[string]float64{"balance": acc.Balance})
			return
		}
	}
	http.Error(w, "Account not found", http.StatusNotFound)
}

func depositBalance(w http.ResponseWriter, r *http.Request, claims *Claims) {
	var body struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// added this check to prevent BOLA- without this validation a user can access data that belongs to another user (unauthorized access)
	if !isUserIdValid(claims, body.UserID) {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	//input validation-checks if the deposit amount is positive
	if body.Amount < 0 {
		http.Error(w, "Illigale amount", http.StatusBadRequest)
		return
	}

	for i, acc := range accounts {
		if acc.UserID == body.UserID {
			accounts[i].Balance += body.Amount
			json.NewEncoder(w).Encode(accounts[i])
			return
		}
	}
	http.Error(w, "Account not found", http.StatusNotFound)
}

func withdrawBalance(w http.ResponseWriter, r *http.Request, claims *Claims) {
	var body struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// added this check to prevent BOLA- without this validation a user can access data that belongs to another user (unauthorized access)
	if !isUserIdValid(claims, body.UserID) {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	//input validation- checks if the deposit amount is positive
	if body.Amount < 0 {
		http.Error(w, "Illigale amount", http.StatusBadRequest)
		return
	}

	for i, acc := range accounts {
		if acc.UserID == body.UserID {
			if acc.Balance < body.Amount {
				http.Error(w, ErrInsufficientFunds.Error(), http.StatusBadRequest)
				return
			}
			accounts[i].Balance -= body.Amount
			json.NewEncoder(w).Encode(accounts[i])
			return
		}
	}
	http.Error(w, "Account not found", http.StatusNotFound)
}

func Auth(next func(http.ResponseWriter, *http.Request, *Claims)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r, claims)
	}
}
