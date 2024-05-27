package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func initserver() {
	ser := http.NewServeMux()
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

	ser.Handle("/ping", a)
	ser.ServeHTTP(nil, nil)
}

func main2() {
	//发送http请求
	reqBody, _ := json.Marshal(map[string]string{"key1": "val1", "key2": "val2"})

	resp, _ := http.Post(":8091", "application/json", bytes.NewReader(reqBody))
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("resp: %s", respBody)
}
