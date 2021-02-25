package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type userProfile struct {
	Data struct {
		GetUserProfileByUsername struct {
			PublicInfo struct {
				DapperID    string `json:"dapperID"`
				FlowAddress string `json:"flowAddress"`
			} `json:"publicInfo"`
		} `json:"getUserProfileByUsername"`
	} `json:"data"`
}

// GetAddress returns as a string either the FLOW address or the Dapper ID
// corresponding to the TS account name provided
func GetAddress(TSUser string, which string) (string, error) {
	// Query
	requestPayload := `{"operationName":"ProfilePage_getUserProfileByUsername","variables":{"input":{"username":"%s"}},"query":"query ProfilePage_getUserProfileByUsername($input: getUserProfileByUsernameInput!) {\n  getUserProfileByUsername(input: $input) {\n    publicInfo {\n      ...UserFragment\n    }\n  }\n}\n\nfragment UserFragment on UserPublicInfo {\n  dapperID\n  flowAddress\n}\n"}`
	body := strings.NewReader(fmt.Sprintf(requestPayload, TSUser))

	resp, err := http.DefaultClient.Post("https://api.nba.dapperlabs.com/marketplace/graphql", "application/json", body)
	if err != nil {
		fmt.Println(err)
	}
	t, _ := ioutil.ReadAll(resp.Body)

	var profile userProfile
	err = json.Unmarshal(t, &profile)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	if which == "Flow" {
		return profile.Data.GetUserProfileByUsername.PublicInfo.FlowAddress, nil
	} else if which == "Dapper" {
		return strings.TrimPrefix(profile.Data.GetUserProfileByUsername.PublicInfo.DapperID, "auth0|"), nil
	} else {
		return "", errors.New("Invalid which parameter")
	}
}

// GetPlayIDFromURL needs fixing
func GetPlayIDFromURL(url string) (string, error) {
	resp, _ := http.Get(url)
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("Error reading HTML body")
	}

	r, _ := regexp.Compile("flowSerialNumber")
	sb := string(body)

	loc := r.FindStringIndex(sb)

	if loc == nil {
		return "", fmt.Errorf("Play ID not detected")
	}

	return sb[loc[1]+3 : loc[1]+13], nil
}
