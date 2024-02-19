package handlers

import "net/http"

func PostInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//body := r.Body
}
