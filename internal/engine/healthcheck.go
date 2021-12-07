package engine

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func waitForAvailableViaHttp(address string, port int) error {

	for i := 0; i < 10; i++ {
		url := fmt.Sprintf("%s://%s:%d%s", "http", address, port, "/")
		httpReq, err := http.NewRequest("GET", url, nil)
		if err == nil {
			httpClient := http.Client{}
			resp, err := httpClient.Do(httpReq)
			if err == nil {
				defer resp.Body.Close()

				ioutil.ReadAll(resp.Body)

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					return nil
				}
			}
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("not available")
}