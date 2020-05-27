package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"

	"github.com/philips-labs/dct-notary-admin/lib/notary"
	"github.com/philips-labs/dct-notary-admin/lib/targets"
)

const (
	appName           = "dctna"
	defaultServerAddr = "https://localhost:8443"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.HelpName = appName
	app.Usage = "Retrieve Docker Content Trust Certificates"
	app.Version = fmt.Sprintf("%s, %s %s", version, commit, date)

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s version %s %s/%s\n", appName, app.Version, runtime.GOOS, runtime.GOARCH)
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "server-address",
			Usage:       "`ENDPOINT` for dctna-server",
			DefaultText: defaultServerAddr,
			EnvVars:     []string{"DCTNA_SERVER_ADDR"},
		},
	}
	app.Action = Run
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Run executes the cli command
func Run(c *cli.Context) error {
	args := c.Args()
	serverAddr := c.String("server-address")
	if serverAddr == "" {
		serverAddr = defaultServerAddr
	}

	if args.Len() == 0 {
		err := errors.New("Please provide one or more targets")
		return err
	}

	keys, err := fetchKeys(serverAddr, args.Slice())
	if err != nil {
		return err
	}

	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	keydir := filepath.Join(home, ".docker", "trust", "private")
	fmt.Println("Saving keys to: ", keydir)
	err = storeKeys(keydir, keys)

	return err
}

func storeKeys(keydir string, keys map[string]notary.KeyData) error {
	for k, v := range keys {
		file := filepath.Join(keydir, k+".key")
		fmt.Printf("Saving Key: %s\n\trole: %s, target: %s\n\n", file, v.Role, v.GUN)
		err := ioutil.WriteFile(file, v.Key, 0600)
		if err != nil {
			return err
		}
	}
	return nil
}

func fetchKeys(serverAddr string, targets []string) (map[string]notary.KeyData, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	keys := make(map[string]notary.KeyData)
	for _, target := range targets {
		keydata, err := fetchTargetKeys(client, serverAddr, target)
		if err != nil {
			return keys, err
		}
		for k, v := range keydata.Data {
			if _, found := keys[k]; !found {
				keys[k] = v
			}
		}
	}
	return keys, nil
}

func fetchTargetKeys(client *http.Client, serverAddr, target string) (*targets.KeyDataResponse, error) {
	jsonData, err := json.Marshal(targets.RepositoryRequest{GUN: target})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, serverAddr+"/api/targets/fetchkeys", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var keydata targets.KeyDataResponse
	err = json.NewDecoder(resp.Body).Decode(&keydata)
	if err != nil {
		return nil, err
	}

	return &keydata, nil
}
