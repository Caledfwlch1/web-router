// rout
package main

import (
	"fmt"
	"net/http"
	//"net"
	"encoding/json"
	"io/ioutil"
	"net/http/httputil"
	"net/url"
	//"strings"

	"github.com/gondor/depcon/marathon"
)

const (
	APP1 = "andyapp1"
	APP2 = "andyapp2"
)

type balanceType struct {
	HostsApp map[string][]string
	CurIndex int
}

func (b *balanceType) BalanceString(app string) string {
	return b.HostsApp[app][b.balanceIndex(app)]
}

func (b *balanceType) balanceIndex(app string) int {
	b.CurIndex++
	if b.CurIndex > len(b.HostsApp[app])-1 {
		b.CurIndex = 0
	}
	return b.CurIndex
}

func (b balanceType) String() (retString string) {

	for n, m := range b.HostsApp {
		s := ""
		for _, k := range m {
			s += k + ", "
		}
		retString += "App: " + n + ", Hosts: " + s
	}
	return retString
}

func (b *balanceType) getTasks(app string, mc marathon.Marathon) {
	tasksApp, _ := mc.GetTasks(app)
	for _, t := range tasksApp {
		b.HostsApp[app] = append(b.HostsApp[app], fmt.Sprintf("http://%s:%d/", t.Host, t.Ports[0]))
	}
}

var hosts balanceType

func main() {
	fmt.Println("Version 9.1")
	hosts.HostsApp = make(map[string][]string)
	ipStr, _ := IP()
	fmt.Println(ipStr)
	handlerMap := make(map[string]func(w http.ResponseWriter, r *http.Request))

	mc := marathon.NewMarathonClient("http://ec2-52-34-228-148.us-west-2.compute.amazonaws.com:8080", "", "")

	hosts.getTasks(APP1, mc)
	hosts.getTasks(APP2, mc)

	fmt.Println(hosts)

	handlerMap["/"] = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "run /")
	}

	handlerMap["/app1"] = func(w http.ResponseWriter, r *http.Request) {
		connectToHost := hosts.BalanceString(APP1)
		if _, err := http.Get(connectToHost); err != nil {
			fmt.Fprintln(w, err)
		} else {
			targetApp, _ := url.Parse(connectToHost)
			revProxyApp := httputil.NewSingleHostReverseProxy(targetApp)
			revProxyApp.ServeHTTP(w, r)
		}
	}

	handlerMap["/app2"] = func(w http.ResponseWriter, r *http.Request) {
		connectToHost := hosts.BalanceString(APP2)
		if _, err := http.Get(connectToHost); err != nil {
			fmt.Fprintln(w, err)
		} else {
			targetApp, _ := url.Parse(connectToHost)
			revProxyApp := httputil.NewSingleHostReverseProxy(targetApp)
			revProxyApp.ServeHTTP(w, r)
		}
	}

	handlerMap["/app2/plus1"] = func(w http.ResponseWriter, r *http.Request) {
		connectToHost := hosts.BalanceString(APP2)
		fmt.Println(connectToHost)
		if resp, err := http.Get(connectToHost); err == nil {
			read, _ := ioutil.ReadAll(resp.Body)
			fmt.Fprintln(w, fmt.Sprintf("%s", read))
		} else {
			fmt.Fprintln(w, err)
		}
		
		connectToHost = hosts.BalanceString(APP1)
		fmt.Println(connectToHost)
		if resp, err := http.Get(connectToHost); err == nil {
			read, _ := ioutil.ReadAll(resp.Body)
			fmt.Fprintln(w, fmt.Sprintf("%s", read))
		} else {
			fmt.Fprintln(w, err)
		}
	}

	for lnk, funcHundler := range handlerMap {
		http.HandleFunc(lnk, funcHundler)
	}

	err := http.ListenAndServe("0.0.0.0:80", nil)
	fmt.Println(err)
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
