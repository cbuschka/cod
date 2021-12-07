package proxy

import (
	"bytes"
	"fmt"
	enginePkg "github.com/cbuschka/cod/internal/engine"
	"io/ioutil"
	"net/http"
)

type Proxy struct {
	engine *enginePkg.Engine
}

func NewProxy(engine *enginePkg.Engine) (*Proxy, error) {
	return &Proxy{engine: engine}, nil
}

func (proxy *Proxy) ForwardToContainer(writer http.ResponseWriter, request *http.Request) error {
	endpoint, err := proxy.engine.GetOrStartContainer("/")
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s://%s:%d%s", "http", endpoint.Address, endpoint.Port, request.RequestURI)

	downstreamReq, err := http.NewRequest(request.Method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	downstreamReq.Header = make(http.Header)
	for h, val := range request.Header {
		downstreamReq.Header[h] = val
	}

	httpClient := http.Client{}
	downstreamResp, err := httpClient.Do(downstreamReq)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadGateway)
		return nil
	}
	defer downstreamResp.Body.Close()

	for key, values := range downstreamResp.Header {
		for _, value := range values {
			writer.Header().Add(key, value)
		}
	}

	downstreamRespBodyBytes, err := ioutil.ReadAll(downstreamResp.Body)
	if err != nil {
		return err
	}
	writer.Write(downstreamRespBodyBytes)

	return nil
}
