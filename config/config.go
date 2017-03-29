/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

// PeerConfig ...
type PeerConfig struct {
	Host                  string
	Port                  string
	EventHost             string
	EventPort             string
	TLSCertificate        string
	TLSServerHostOverride string
}

type fabricCAConfig struct {
	ServerURL string   `json:"serverURL"`
	Certfiles []string `json:"certfiles"`
	Client    struct {
		Keyfile  string `json:"keyfile"`
		Certfile string `json:"certfile"`
	} `json:"client"`
}

var myViper = viper.New()
var log = logging.MustGetLogger("fabric_sdk_go")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} [%{module}] %{level:.4s} : %{color:reset} %{message}`,
)

// InitConfig ...
// initConfig reads in config file
func InitConfig(configFile string) error {

	if configFile != "" {
		// create new viper
		myViper.SetConfigFile(configFile)
		// If a config file is found, read it in.
		err := myViper.ReadInConfig()

		if err == nil {
			log.Infof("Using config file: %s", myViper.ConfigFileUsed())
		} else {
			return fmt.Errorf("Fatal error config file: %v", err)
		}
	}

	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	loggingLevelString := myViper.GetString("client.logging.level")
	logLevel := logging.INFO
	if loggingLevelString != "" {
		log.Infof("fabric_sdk_go Logging level: %v", loggingLevelString)
		var err error
		logLevel, err = logging.LogLevel(loggingLevelString)
		if err != nil {
			panic(err)
		}
	}
	logging.SetBackend(backendFormatter).SetLevel(logging.Level(logLevel), "fabric_sdk_go")

	return nil
}

// GetFabricClientViper returns the internal viper instance used by the
// SDK to read configuration options
func GetFabricClientViper() *viper.Viper {
	return myViper
}

// GetPeersConfig ...
func GetPeersConfig() []PeerConfig {
	peersConfig := []PeerConfig{}
	peers := myViper.GetStringMap("client.peers")
	for key, value := range peers {
		mm, ok := value.(map[string]interface{})
		var host string
		var port int
		var eventHost string
		var eventPort int
		var tlsCertificate string
		var tlsServerHostOverride string
		if ok {
			host, _ = mm["host"].(string)
			port, _ = mm["port"].(int)
			eventHost, _ = mm["event_host"].(string)
			eventPort, _ = mm["event_port"].(int)
			tlsCertificate, _ = mm["tls"].(map[string]interface{})["certificate"].(string)
			tlsServerHostOverride, _ = mm["tls"].(map[string]interface{})["serverhostoverride"].(string)

		} else {
			mm1 := value.(map[interface{}]interface{})
			host, _ = mm1["host"].(string)
			port, _ = mm1["port"].(int)
			eventHost, _ = mm1["event_host"].(string)
			eventPort, _ = mm1["event_port"].(int)
			tlsCertificate, _ = mm1["tls"].(map[string]interface{})["certificate"].(string)
			tlsServerHostOverride, _ = mm1["tls"].(map[string]interface{})["serverhostoverride"].(string)

		}

		p := PeerConfig{Host: host, Port: strconv.Itoa(port), EventHost: eventHost, EventPort: strconv.Itoa(eventPort),
			TLSCertificate: tlsCertificate, TLSServerHostOverride: tlsServerHostOverride}
		if p.Host == "" {
			panic(fmt.Sprintf("host key not exist or empty for %s", key))
		}
		if p.Port == "" {
			panic(fmt.Sprintf("port key not exist or empty for %s", key))
		}
		if p.EventHost == "" {
			panic(fmt.Sprintf("event_host not exist or empty for %s", key))
		}
		if p.EventPort == "" {
			panic(fmt.Sprintf("event_port not exist or empty for %s", key))
		}
		if IsTLSEnabled() && p.TLSCertificate == "" {
			panic(fmt.Sprintf("tls.certificate not exist or empty for %s", key))
		}

		p.TLSCertificate = strings.Replace(p.TLSCertificate, "$GOPATH", os.Getenv("GOPATH"), -1)
		peersConfig = append(peersConfig, p)
	}
	return peersConfig

}

// IsTLSEnabled ...
func IsTLSEnabled() bool {
	return myViper.GetBool("client.tls.enabled")
}

// GetTLSCACertPool ...
func GetTLSCACertPool(tlsCertificate string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	if tlsCertificate != "" {
		rawData, err := ioutil.ReadFile(tlsCertificate)
		if err != nil {
			return nil, err
		}

		certPool.AddCert(loadCAKey(rawData))
	}

	return certPool, nil
}

// IsSecurityEnabled ...
func IsSecurityEnabled() bool {
	return myViper.GetBool("client.security.enabled")
}

// TcertBatchSize ...
func TcertBatchSize() int {
	return myViper.GetInt("client.tcert.batch.size")
}

// GetSecurityAlgorithm ...
func GetSecurityAlgorithm() string {
	return myViper.GetString("client.security.hashAlgorithm")
}

// GetSecurityLevel ...
func GetSecurityLevel() int {
	return myViper.GetInt("client.security.level")

}

// GetOrdererHost ...
func GetOrdererHost() string {
	return myViper.GetString("client.orderer.host")
}

// GetOrdererPort ...
func GetOrdererPort() string {
	return strconv.Itoa(myViper.GetInt("client.orderer.port"))
}

// GetOrdererTLSServerHostOverride ...
func GetOrdererTLSServerHostOverride() string {
	return myViper.GetString("client.orderer.tls.serverhostoverride")
}

// GetOrdererTLSCertificate ...
func GetOrdererTLSCertificate() string {
	return strings.Replace(myViper.GetString("client.orderer.tls.certificate"), "$GOPATH", os.Getenv("GOPATH"), -1)
}

// GetFabricCAID ...
func GetFabricCAID() string {
	return myViper.GetString("client.fabricCA.id")
}

// GetFabricCAClientPath This method will read the fabric-ca configurations from the
// config yaml file and return the path to a json client config file
// in the format that is expected by the fabric-ca client
func GetFabricCAClientPath() (string, error) {
	filePath := "/tmp/client-config.json"
	fabricCAConf := fabricCAConfig{}
	err := myViper.UnmarshalKey("client.fabricCA", &fabricCAConf)
	if err != nil {
		return "", err
	}
	jsonConfig, err := json.Marshal(fabricCAConf)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filePath, jsonConfig, 0644)
	return filePath, err
}

// GetKeyStorePath ...
func GetKeyStorePath() string {
	return myViper.GetString("client.keystore.path")
}

// loadCAKey
func loadCAKey(rawData []byte) *x509.Certificate {
	block, _ := pem.Decode(rawData)

	pub, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}
	return pub
}