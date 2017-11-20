package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// User represents the User API method handler set.
type User struct {
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing users in the system.
func (u *User) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	data := struct {
		Name  string
		Email string
	}{
		Name:  "Bill",
		Email: "bill@ardanlabs.com",
	}

	// Marshal the data into a JSON string.
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("Respond %v Marshalling JSON response\n", err)
		jsonData = []byte("{}")
	}

	// Send the result back to the client.
	io.WriteString(w, string(jsonData))
	return nil
}
