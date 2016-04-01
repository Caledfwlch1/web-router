// app1
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

func main() {

	ipStr, _ := IP()
	fmt.Println(ipStr)

	handlerMap := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Run App1.")

		printInterfaces(w)
	}
	http.HandleFunc("/", handlerMap)
	err := http.ListenAndServe("0.0.0.0:81", nil)
	fmt.Println(err)
}

func printInterfaces(w http.ResponseWriter) {
	n, _ := net.InterfaceAddrs()
	for _, i := range n {
		fmt.Fprintln(w, i.String())
	}
}

func IP() (string, error) {
	r, err := http.Get("http://api.ipify.org/?format=json")
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	info := struct {
		Ip string `json:"ip"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		return "", err
	}
	return info.Ip, nil
}
