package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	//initserver()
	var a http.HandlerFunc
	a = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		for i := 0; i <= 10; i++ {
			w.Write([]byte(fmt.Sprintf("pong %d \n", i)))
			w.(http.Flusher).Flush()
			//time.Sleep(1 * time.Second)
		}
		//w.(http.CloseNotifier).CloseNotify()

		//a, _ := json.Marshal(r.Context().Value(http.ServerContextKey).(*http.Server))
		b, _ := json.Marshal(r.RemoteAddr)
		//w.Write(a)
		w.Write(b)
	}
	http.HandleFunc("/", a)
	w := httptest.NewRecorder()
	//reqbody := map[string]string{
	//date, _ := json.Marshal(reqbody)
	//req := httptest.NewRequest("POST", "/api/generate/user", bytes.NewReader(date))
	req := httptest.NewRequest("Post", "/", nil)
	a.ServeHTTP(w, req)
	fmt.Println(w.Body.String())
}
