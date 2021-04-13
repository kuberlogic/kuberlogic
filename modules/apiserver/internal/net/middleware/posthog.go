package middleware

import (
	"fmt"
	"net/http"
)

func PrintValues(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		fmt.Println("method ===>", r.Method)
		fmt.Println("url ===>", r.URL)
		fmt.Println("form ===>", r.Form)
		fmt.Println("yaform ===>", r.PostForm)
	})
}
