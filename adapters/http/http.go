package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// NewServer creates a new http handler
func NewServer(adapter BoxHTTPAdapter) http.Handler {

	r := mux.NewRouter()
	r.HandleFunc("/orders", PostOrdersHandler(adapter)).Methods("POST")
	r.HandleFunc("/status/{orderId}", GetStatusHandler(adapter)).Methods("GET")
	r.HandleFunc("/whoami/{challenge}", WhoAmIHandler(adapter)).Methods("GET")
	r.HandleFunc("/balances", GetBalancesHandler(adapter)).Methods("GET")
	r.Use(RecoveryHandler)

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST"},
	}).Handler(r)
	return handler
}

// RecoveryHandler handles errors while processing the requests and populates the errors in the response
func RecoveryHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				writeError(w, http.StatusInternalServerError, fmt.Sprintf("%v", r))
			}
		}()
		h.ServeHTTP(w, r)
	})
}

// PostOrdersHandler handles post orders request, it gets the signed order id,
// checks whether the signer is authorized, if the signer is authorized this
// function adds the order id to the queue.
func PostOrdersHandler(boxHTTPAdapter BoxHTTPAdapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postOrder := PostOrder{}
		if err := json.NewDecoder(r.Body).Decode(&postOrder); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("cannot decode json into post order format: %v", err))
			return
		}

		processedOrder, err := boxHTTPAdapter.PostOrder(postOrder)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("cannot process the order: %v", err))
			return
		}

		orderJSON, err := json.Marshal(processedOrder)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("cannot marshal the processed order: %v", err))
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(orderJSON)
	}
}

// WhoAmIHandler handles the get whoami request,it gets a challenge from the
// caller signs it and sends back the signed challenge with it's version
// information.
func WhoAmIHandler(adapter BoxHTTPAdapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		whoami, err := adapter.WhoAmI(params["challenge"])
		if err != nil {

			writeError(w, http.StatusInternalServerError, fmt.Sprintf("cannot get the whoami information: %v", err))
			return
		}
		whoamiJSON, err := json.Marshal(whoami)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("cannot marshal whoami information: %v", err))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(whoamiJSON)
	}
}

//
func GetStatusHandler(adapter BoxHTTPAdapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		status, err := adapter.GetStatus(params["orderId"])
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("cannot get the status information: %v", err))
			return
		}

		statusJSON, err := json.Marshal(status)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("cannot marshal status information: %v", err))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(statusJSON)
	}
}

func GetBalancesHandler(adapter BoxHTTPAdapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		balances, err := adapter.GetBalances()
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("cannot get the balances: %v", err))
			return
		}

		balancesJSON, err := json.Marshal(balances)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("cannot marshal the balance information: %v", err))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(balancesJSON)
	}
}

func writeError(w http.ResponseWriter, statusCode int, err string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(err))
	return
}
