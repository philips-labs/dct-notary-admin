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
	meta, err := fetchMetadata(serverAddr, args.Slice())
	if err != nil {
		return err
	}

	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	trustdir := filepath.Join(home, ".docker", "trust")
	keydir := filepath.Join(trustdir, "private")
	tufdir := filepath.Join(trustdir, "tuf")
	fmt.Println("Saving keys to: ", keydir)
	err = storeKeys(keydir, keys)
	if err != nil {
		return err
	}
	fmt.Println("Saving metadata to: ", tufdir)
	err = storeMetadata(tufdir, meta)

	return err
}

func storeMetadata(tufdir string, metadata map[string]notary.TUFMetadata) error {
	for k, v := range metadata {
		metadir := filepath.Join(tufdir, k, "metadata")

		err := os.MkdirAll(metadir, 0700)
		if err != nil {
			return err
		}

		if v.Root != nil {
			rootfile := filepath.Join(metadir, "root.json")
			fmt.Printf("\t%s\n", rootfile)
			json, err := v.Root.MarshalJSON()
			if err != nil {
				return err
			}
			err = writeMetadataJSON(rootfile, json)
			if err != nil {
				return err
			}
		}
		if v.Targets != nil {
			for k, v := range v.Targets {
				err := os.MkdirAll(filepath.Join(metadir, "targets"), 0700)
				if err != nil {
					return err
				}
				targetsfile := filepath.Join(metadir, k.String()+".json")
				fmt.Printf("\t%s\n", targetsfile)
				json, err := v.MarshalJSON()
				if err != nil {
					return err
				}
				err = writeMetadataJSON(targetsfile, json)
				if err != nil {
					return err
				}
			}
		}
		if v.Snapshot != nil {
			snapshotfile := filepath.Join(metadir, "snapshot.json")
			fmt.Printf("\t%s\n", snapshotfile)
			json, err := v.Snapshot.MarshalJSON()
			if err != nil {
				return err
			}
			err = writeMetadataJSON(snapshotfile, json)
			if err != nil {
				return err
			}
		}
		if v.Timestamp != nil {
			timestampfile := filepath.Join(metadir, "timestamp.json")
			fmt.Printf("\t%s\n", timestampfile)
			json, err := v.Timestamp.MarshalJSON()
			if err != nil {
				return err
			}
			err = writeMetadataJSON(timestampfile, json)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func writeMetadataJSON(file string, json []byte) error {
	return ioutil.WriteFile(file, json, 0600)
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

func fetchMetadata(serverAddr string, targets []string) (map[string]notary.TUFMetadata, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	meta := make(map[string]notary.TUFMetadata)
	for _, target := range targets {
		metadata, err := fetchTargetMetadata(client, serverAddr, target)
		if err != nil {
			return meta, err
		}
		if metadata.Data != nil {
			meta[target] = *metadata.Data
		}
	}

	return meta, nil
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

func fetchFromAPI(client *http.Client, addr, target string, respBody interface{}) error {
	jsonData, err := json.Marshal(targets.RepositoryRequest{GUN: target})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, addr, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(respBody)
	if err != nil {
		return err
	}

	return nil
}

func fetchTargetKeys(client *http.Client, serverAddr, target string) (*targets.KeyDataResponse, error) {
	var keydata targets.KeyDataResponse
	err := fetchFromAPI(client, serverAddr+"/api/targets/fetchkeys", target, &keydata)
	if err != nil {
		return nil, err
	}

	return &keydata, nil
}

func fetchTargetMetadata(client *http.Client, serverAddr, target string) (*targets.MetadataResponse, error) {
	var metadata targets.MetadataResponse
	err := fetchFromAPI(client, serverAddr+"/api/targets/fetchmeta", target, &metadata)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}
