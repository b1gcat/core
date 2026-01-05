package machineid

import (
	"io"
	"net/http"
)

// curl https://ipinfo.io/json
func GetIP(token string) ([]byte, error) {
	resp, err := http.Get("https://api.ipinfo.io/lite/me?token=" + token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
