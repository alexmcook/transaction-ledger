package api

import (
	"fmt"
	"net/http"
	"strings"
	"strconv"
	"github.com/alexmcook/transaction-ledger/internal/service"
)

func NewRouter(svc *service.Service) http.Handler {
	mux := http.NewServeMux()
	registerRoutes(mux, svc)
	return mux
}

func registerRoutes(mux *http.ServeMux, svc *service.Service) {
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/users", handleUsers(svc))
	mux.HandleFunc("/users/", handleUsers(svc))
	mux.HandleFunc("/accounts", handleAccounts(svc))
}

// @Summary API Health check
// @Description Returns 200 OK if the API is running
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /health [get]
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func handleUsers(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(p, "/")
		switch r.Method {
		case http.MethodGet:
			if len(parts) != 2 {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// Extract user ID from URL
			r.SetPathValue("userId", parts[1])
		  handleGetUser(svc)(w,r)
			return
		case http.MethodPost:
			handleCreateUser(svc)(w, r)
			return
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
	}
}

// @Summary Get user by ID
// @Description Retrieves a user by their ID
// @Produce plain
// @Param id path int true "User ID"
// @Success 200 {string} string "User ID"
// @Failure 400 {string} string "Invalid user ID"
// @Failure 404 {string} string "User not found"
// @Router /users/{id} [get]
func handleGetUser(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from URL
		userIdStr := r.PathValue("userId")
		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			fmt.Fprintf(w, "Err: %s", userIdStr)
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		user, err := svc.Users.GetUser(r.Context(), userId)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "User ID: %d", user.Id)
	}
}

// @Summary Create a new user
// @Description Creates a new user in the system
// @Produce plain
// @Success 201 {string} string "User created with ID"
// @Failure 500 {string} string "Failed to create user"
// @Router /users [post]
func handleCreateUser(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := svc.Users.CreateUser(r.Context())
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "User created with ID: %d", user.Id)
	}
}

// @Summary Create a new account
// @Description Creates a new account for a user
// @Produce plain
// @Success 201 {string} string "Account created with ID"
// @Failure 500 {string} string "Failed to create account"
// @Router /accounts [post]
func handleAccounts(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, err := svc.Accounts.CreateAccount(r.Context(), 1, 1000)
		if err != nil {
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Account created with ID: %d", account.Id)
	}
}
