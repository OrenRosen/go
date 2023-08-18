package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func main() {
	s := NewService()

	http.Handle("/do-work", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i, err := strconv.Atoi(r.URL.Query().Get("i"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "i must be int")
			return
		}

		s.DoWorkAsync(r.Context(), i)

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "working...")
	}))

	fmt.Println("Listening...")
	fmt.Println("for work, go to: http://127.0.0.1:4684/do-work?i=10")

	http.ListenAndServe(":4684", nil)
}
