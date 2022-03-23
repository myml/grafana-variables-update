package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	token := os.Getenv("GRAFANA_TOKEN")
	host := os.Getenv("GRAFANA_HOST")
	dashboard := os.Getenv("GRAFANA_DASHBOARD")
	orgName := os.Getenv("ORG_NAME")
	orgMembers := os.Getenv("ORG_MEMBERS")

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, host+"/api/dashboards/uid/"+dashboard, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	path := `dashboard.templating.list.#(name="org").options.#(text="` + orgName + `").value`
	if !gjson.GetBytes(data, path).Exists() {
		log.Fatal("path not exists")
	}
	data, err = sjson.SetBytes(data, path, orgMembers)
	if err != nil {
		log.Fatal(err)
	}
	path = `dashboard.templating.list.#(name="org").options.#(text="Other").value`
	if !gjson.GetBytes(data, path).Exists() {
		log.Fatal("path not exists")
	}
	data, err = sjson.SetBytes(data, path, fmt.Sprintf("NOT (%s)", orgMembers))
	if err != nil {
		log.Fatal(err)
	}
	path = `dashboard.templating.list.#(name="org").query`
	data, err = sjson.SetBytes(data, path, fmt.Sprintf("%s : %s, Other : NOT (%s)", orgName, orgMembers, orgMembers))
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, host+"/api/dashboards/db/", bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stderr, resp.Body)
		log.Fatal(resp.Status)
	}
}
