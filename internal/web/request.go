package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

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
	err = json.Unmarshal(resBody, &resComplex)
	if err == nil {
		return resComplex.Error.Message
	}
	var resDesc responseErrorDescribed
	err = json.Unmarshal(resBody, &resDesc)
	if err == nil {
		return resDesc.ErrorDescription
	}
	var resSimple responseErrorSimple
	err = json.Unmarshal(resBody, &resSimple)
	if err == nil {
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

func DispatchRequest(req *http.Request, output interface{}) (int, error) {
	client := &http.Client{}
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
	client := &http.Client{}
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
