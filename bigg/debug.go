package bigg

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

type loggingRoundTripper struct {
	proxy http.RoundTripper
}

func (lrt loggingRoundTripper) RoundTrip(req *http.Request) (res *http.Response, e error) {
	dump, _ := httputil.DumpRequestOut(req, true)
	fmt.Println("\n___________________________________________________________")
	fmt.Printf("[REQUEST]\n%s\n", string(dump))
	res, e = lrt.proxy.RoundTrip(req)
	if e != nil {
		fmt.Printf("_______\n[ERROR]\n%v\n", e)
	} else {
		dump, _ := httputil.DumpResponse(res, true)
		fmt.Printf("__________\n[RESPONSE]\n%s\n", string(dump))
	}
	fmt.Print("___________________________________________________________\n\n")
	return
}
