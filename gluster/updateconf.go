package main

import (
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type cVar struct {
	Name     string   `json:"name,omitempty"`
	Settable []string `json:"settable,omitempty"`
	Value    string   `json:"value,omitempty"`
}

type cArgs struct {
	Description string   `json:"Description,omitempty"`
	Name        string   `json:"Name,omitempty"`
	Settable    []string `json:"Settable,omitempty"`
	Value       []string `json:"Value,omitempty"`
}

type cInterface struct {
	Socket string   `json:"socket,omitempty"`
	Types  []string `json:"types,omitempty"`
}

type cDevice struct {
	Path string `json:"path,omitempty"`
}

type cLinux struct {
	Capabilities []string  `json:"capabilities,omitempty"`
	Devices      []cDevice `json:"devices,omitempty"`
}

type cType struct {
	Type string `json:"type,omitempty"`
}

type config struct {
	Description   string     `json:"description,omitempty"`
	Documentation string     `json:"documentation,omitempty"`
	Entrypoint    []string   `json:"entrypoint,omitempty"`
	Env           []cVar     `json:"env,omitempty"`
	Args          cArgs      `json:"Args,omitempty"`
	Interface     cInterface `json:"interface,omitempty"`
	Linux         cLinux     `json:"linux,omitempty"`
	Network       cType      `json:"network,omitempty"`
	Mount         string     `json:"propagatedmount,omitempty"`
}

func readFileToCVar(key, name string) cVar {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Println("Error opening File", err)
	}

	return kvToCVar(key, string(data))
}

// kvToCVar takes in keys and value and returns a cVar object
func kvToCVar(key, value string) cVar {

	return cVar{
		Name:     key,
		Settable: []string{"value"},
		Value:    value,
	}
}

func loadCredsToConfig(c *config) error {
	credsDir, ok := os.LookupEnv("CREDS_DIR")
	if !ok {
		return fmt.Errorf("CREDS_DIR is empty")
	}

	credsFile := credsDir + "/" + "creds.txt"
	data, err := ioutil.ReadFile(credsFile)
	if err != nil {
		fmt.Println(err)
		return err
	}

	content := strings.Split(string(data), "\n")
	if len(content) > 2 {
		user := content[0]
		password := content[1]

		c.Env = append(c.Env, kvToCVar("user", user))
		c.Env = append(c.Env, kvToCVar("password", password))
	}
	return nil
}

func loadPortToConfig(c *config) {
	port, _ := os.LookupEnv("PORT")

	if len(port) > 0 {
		c.Env = append(c.Env, kvToCVar("port", port))
	}
}

func loadKeysToConfig(c *config) error {
	// Read TLS_KEY
	file, _ := os.LookupEnv("TLS_KEY")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	ok := isValidPEM(data)
	if !ok {
		fmt.Println("Wrong PEM data")
		return fmt.Errorf("key not in pem format")
	}

	c.Env = append(c.Env, kvToCVar("key", string(data)))

	file, _ = os.LookupEnv("TLS_CERT")
	data, err = ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	ok = isValidPEM(data)
	if !ok {
		fmt.Println("Wrong PEM data")
		return fmt.Errorf("cert not in pem format")
	}

	c.Env = append(c.Env, kvToCVar("cert", string(data)))
	file, _ = os.LookupEnv("TLS_CACERT")
	data, err = ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	ok = isValidPEM(data)
	if !ok {
		fmt.Println("Wrong PEM data")
		return fmt.Errorf("cacert not in pem format")
	}

	c.Env = append(c.Env, kvToCVar("cacert", string(data)))
	return nil
}

func dumpConfig(c config) {
	jData, err := json.Marshal(c)
	if err != nil {
		return
	}
	ioutil.WriteFile("config.json", jData, 0664)
}

func isValidPEM(data []byte) bool {
	block, _ := pem.Decode(data)
	if block == nil {
		return false
	}

	if block.Type == "RSA PRIVATE KEY" {
		return true
	}

	if block.Type == "CERTIFICATE" {
		return true
	}

	return false
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: updateconf /path/to/config.json")
		return
	}

	configFilePath := os.Args[1]

	if len(configFilePath) == 0 {
		fmt.Println("Error opening File")
		return
	}

	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		fmt.Println("Error opening File", err.Error())
		return
	}

	if len(data) == 0 {
		fmt.Println("Error opening File")
		return
	}

	var DJ config
	json.Unmarshal(data, &DJ)

	err = loadCredsToConfig(&DJ)
	if err != nil {
		panic("No credentials found")
	}

	err = loadKeysToConfig(&DJ)
	if err != nil {
		panic("No proper keys found")
	}
	loadPortToConfig(&DJ)

	/*
	   It reads the following variables and appends them to the config in the
	   required format
	    - PORT
	    - TLS_KEY
	    - TLS_CERT
	    - TLS_CACERT
	    - CREDS_DIR

	     CREDS_DIR contains a file `creds.txt` in which first line has to be takes as user
	     and the second as password.
	*/

	dumpConfig(DJ)
}
