package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var CommentClient commentClient

type commentClient struct {
	Client  http.Client
	baseUrl string
	token   string
}

func (c *commentClient) getToken() error {

	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/Token/GetToken?apikey="+os.Getenv("apikey"), nil)
	if err != nil {
		return err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	b := make(map[string]string)
	err = json.Unmarshal(respBody, &b)
	if err != nil {
		return nil
	}

	token := b["token"]
	c.token = token

	return nil

}

func (c *commentClient) swearWordTagger(comment string) (map[string]string, error) {

	reqJson, err := json.Marshal(comment)
	if err != nil {
		return nil, err
	}

	reqBody := bytes.NewReader(reqJson)

	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/TextRefinement/SwearWordTagger", reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	var resp *http.Response

	i := 0
	for i < 3 {
		req.Header.Del("Authorization")
		req.Header.Add("Authorization", "Bearer "+c.token)
		resp, err = c.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			break
		} else if resp.StatusCode == http.StatusUnauthorized {
			c.getToken()
		}

		i++
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	b := make(map[string]string)
	err = json.Unmarshal(respBody, &b)
	if err != nil {
		return nil, err
	}

	return b, nil

}

func (c *commentClient) IsSafe(comment string) (bool, error) {
	swears, err := CommentClient.swearWordTagger(comment)
	if err != nil {
		return false, err
	}

	mildCount := 0
	for _, v := range swears {
		if v == "StrongSwearWord" {
			return false, nil
		} else if v == "MildSwearWord" {
			mildCount++
		}
		if mildCount > 1 {
			return false, nil
		}
	}

	return true, nil
}

func init() {

	CommentClient.Client.Timeout = time.Second * 5
	CommentClient.baseUrl = "https://api.text-mining.ir/api"
	CommentClient.getToken()

}
