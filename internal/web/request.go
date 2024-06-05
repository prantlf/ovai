package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/prantlf/ovai/internal/log"
)

type errorObj struct {
	Message string `json:"message"`
}

type responseErrorComplex struct {
	Error errorObj `json:"error"`
}

type responseErrorDescribed struct {
	ErrorDescription string `json:"error_description"`
}

type responseErrorSimple struct {
	Error string `json:"error"`
}

func logError(req *http.Request, status int, resBody []byte) {
	if log.IsDbg {
		var resLog bytes.Buffer
		if errLog := json.Indent(&resLog, resBody, "", "  "); errLog != nil {
			log.Dbg("receive %d: %s %s\n with response %s", status, req.Method, req.URL, resBody)
			// log.Printf("receive %d: %s %s\n with response %+v", status, req.Method, req.URL, resOutput)
		} else {
			log.Dbg("receive %d: %s %s\n with response %s", status, req.Method, req.URL, resLog.Bytes())
		}
	}
}

func readError(req *http.Request, res *http.Response) string {
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Dbg("receive %d: %s %s", res.StatusCode, req.Method, req.URL)
		return ""
	}
	logError(req, res.StatusCode, resBody)
	var resComplex responseErrorComplex
	if err = json.Unmarshal(resBody, &resComplex); err == nil && len(resComplex.Error.Message) > 0 {
		return resComplex.Error.Message
	}
	var resDesc responseErrorDescribed
	if err = json.Unmarshal(resBody, &resDesc); err == nil && len(resDesc.ErrorDescription) > 0 {
		return resDesc.ErrorDescription
	}
	var resSimple responseErrorSimple
	if err = json.Unmarshal(resBody, &resSimple); err == nil && len(resSimple.Error) > 0 {
		return resSimple.Error
	}
	return string(resBody)
}

func CreatePostRequest(url string, input interface{}) (*http.Request, error) {
	if log.IsNet {
		inputJson, errLog := json.MarshalIndent(input, "", "  ")
		if errLog != nil {
			log.Net("send POST %s\n with body %+v", url, input)
		} else {
			log.Net("send POST %s\n with body %s", url, inputJson)
		}
	}
	reader, writer := io.Pipe()
	go func() {
		if err := json.NewEncoder(writer).Encode(input); err != nil {
			log.Dbg("! encoding request body failed: %v", err)
		}
		if err := writer.Close(); err != nil {
			log.Dbg("! closing writer to request pipe failed: %v", err)
		}
	}()

	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Dbg("preparing request failed: %v", err)
		return nil, errors.New("preparing request failed")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func CreateGetRequest(url string) (*http.Request, error) {
	log.Net("send GET %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Dbg("preparing request failed: %v", err)
		return nil, errors.New("preparing request failed")
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func CreateRawPostRequest(url string, input []byte) (*http.Request, error) {
	log.Net("send POST %s\n with body %s", url, input)
	reader, writer := io.Pipe()
	go func() {
		if _, err := writer.Write(input); err != nil {
			log.Dbg("! writing request body failed: %v", err)
		}
		if err := writer.Close(); err != nil {
			log.Dbg("! closing writer to request pipe failed: %v", err)
		}
	}()
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Dbg("preparing request failed: %v", err)
		return nil, errors.New("preparing request failed")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// type happyEyeballs int

// const (
// 	ipDefault happyEyeballs = iota + 1
// 	ipV4
// 	ipV6
// )

// var happyEyeballNames = [...]string{"Default", "IPV4", "IPV6"}

// func (h happyEyeballs) String() string {
// 	return happyEyeballNames[h-1]
// }

// func (h happyEyeballs) EnumIndex() int {
// 	return int(h)
// }

// func parseHappyEyeballs(input string) (happyEyeballs, error) {
// 	for index, value := range happyEyeballNames {
// 		if value == input {
// 			return happyEyeballs(index), nil
// 		}
// 	}
// 	return 0, fmt.Errorf("invalid enum value: %q", input)
// }

var dialer net.Dialer
var networkVersion = initDialer()

func initDialer() string {
	network := os.Getenv("NETWORK")
	if len(network) > 0 {
		if network == "IPV4" {
			return "tcp4"
		}
		if network == "IPV6" {
			return "tcp6"
		}
		log.Ftl("Invalid value of NETWORK variable: %q", network)
	}
	return ""
}

func createHttpClient() *http.Client {
	client := &http.Client{}
	if len(networkVersion) > 0 {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, networkVersion, addr)
		}
		client.Transport = transport
	}
	return client
}

func DispatchRequest(req *http.Request, output interface{}) (int, error) {
	client := createHttpClient()
	res, err := client.Do(req)
	if err != nil {
		log.Dbg("making request failed: %v", err)
		return http.StatusInternalServerError, errors.New("making request failed")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Dbg("closing request body stream failed: %v", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		msg := readError(req, res)
		if len(msg) == 0 {
			msg = res.Status
		}
		return res.StatusCode, errors.New(msg)
	}

	// err = json.NewDecoder(res.Body).Decode(output)
	// if err != nil {
	// 	return http.StatusInternalServerError, fmt.Errorf("decoding response body failed: %w", err)
	// }
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Dbg("receive %d: %s %s", res.StatusCode, req.Method, req.URL)
		log.Dbg("reading response body failed: %v", err)
		return http.StatusInternalServerError, errors.New("reading response body failed")
	}
	if err = json.Unmarshal(resBody, output); err != nil {
		log.Dbg("receive %d: %s %s\n with response %s", res.StatusCode, req.Method, req.URL, resBody)
		log.Dbg("decoding response body failed: %v", err)
		return http.StatusInternalServerError, errors.New("decoding response body failed")
	}
	if log.IsDbg {
		var resLog bytes.Buffer
		if errLog := json.Indent(&resLog, resBody, "", "  "); errLog != nil {
			log.Net("receive %d: %s %s\n with response %s", res.StatusCode, req.Method, req.URL, resBody)
			// log.Printf("receive %d: %s %s\n with response %+v", res.StatusCode, req.Method, req.URL, output)
		} else {
			log.Net("receive %d: %s %s\n with response %s", res.StatusCode, req.Method, req.URL, resLog.Bytes())
		}
	}
	return http.StatusOK, nil
}

func DispatchRawRequest(req *http.Request) (int, []byte, error) {
	client := createHttpClient()
	res, err := client.Do(req)
	if err != nil {
		log.Dbg("making request failed: %v", err)
		return http.StatusInternalServerError, nil, errors.New("making request failed")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Dbg("closing request body stream failed: %v", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		msg := readError(req, res)
		if len(msg) == 0 {
			msg = res.Status
		}
		return res.StatusCode, nil, errors.New(msg)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Dbg("receive %d: %s %s", res.StatusCode, req.Method, req.URL)
		log.Dbg("reading response body failed: %v", err)
		return http.StatusInternalServerError, nil, errors.New("reading response body failed")
	}
	if log.IsDbg {
		var resLog bytes.Buffer
		if errLog := json.Indent(&resLog, resBody, "", "  "); errLog != nil {
			log.Net("receive %d: %s %s\n with response %s", res.StatusCode, req.Method, req.URL, resBody)
		} else {
			log.Net("receive %d: %s %s\n with response %s", res.StatusCode, req.Method, req.URL, resLog.Bytes())
		}
	}
	return http.StatusOK, resBody, nil
}
