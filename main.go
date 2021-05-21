package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/deyaeddin/cert-manager-webhook-hetzner/internal"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var GroupName = os.Getenv("GROUP_NAME")
var zapLogger, _ = zap.NewProduction()

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	cmd.RunWebhookServer(GroupName,
		&hetznerDNSProviderSolver{},
	)
}

type hetznerDNSProviderSolver struct {
	client *kubernetes.Clientset
}

type hetznerDNSProviderConfig struct {
	SecretRef string `json:"secretName"`
	ZoneName  string `json:"zoneName"`
	ApiUrl    string `json:"apiUrl"`
}

func (c *hetznerDNSProviderSolver) Name() string {
	return "hetzner"
}

func (c *hetznerDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	slogger := zapLogger.Sugar()
	slogger.Infof("call function Present: namespace=%s, zone=%s, fqdn=%s", ch.ResourceNamespace, ch.ResolvedZone, ch.ResolvedFQDN)

	config, err := clientConfig(c, ch)

	if err != nil {
		return fmt.Errorf("unable to get secret `%s`; %v", ch.ResourceNamespace, err)
	}

	addTxtRecord(config, ch)

	slogger.Infof("Presented txt record %v", ch.ResolvedFQDN)

	return nil
}

func (c *hetznerDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	slogger := zapLogger.Sugar()

	config, err := clientConfig(c, ch)

	if err != nil {
		return fmt.Errorf("unable to get secret `%s`; %v", ch.ResourceNamespace, err)
	}

	zoneId, err := searchZoneId(config)

	if err != nil {
		return fmt.Errorf("unable to find id for zone name `%s`; %v", config.ZoneName, err)
	}

	var url = config.ApiUrl + "/records?zone_id=" + zoneId

	// Get all DNS records
	dnsRecords, err := callDnsApi(url, "GET", nil, config)

	if err != nil {
		return fmt.Errorf("unable to get DNS records %v", err)
	}

	// Unmarshall response
	records := internal.RecordResponse{}
	readErr := json.Unmarshal(dnsRecords, &records)

	if readErr != nil {
		return fmt.Errorf("unable to unmarshal response %v", readErr)
	}

	var recordId string
	name := recordName(ch.ResolvedFQDN, config.ZoneName)
	slogger.Infof("record name after recordName: %s  for FQDN: %s", name, ch.ResolvedFQDN)

	for i := len(records.Records) - 1; i >= 0; i-- {
		if records.Records[i].Name == strings.ToLower(name) {
			recordId = records.Records[i].Id
			break
		}
	}

	slogger.Infof(" deleting recordID: %s", recordId)
	// Delete TXT record
	url = config.ApiUrl + "/records/" + recordId
	del, err := callDnsApi(url, "DELETE", nil, config)

	if err != nil {
		slogger.Error(err)
	}
	slogger.Infof("Delete TXT record result: %s", string(del))
	return nil
}

func (c *hetznerDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {

	slogger := zapLogger.Sugar()
	k8sClient, err := kubernetes.NewForConfig(kubeClientConfig)
	slogger.Infof("Input variable stopCh is %d length", len(stopCh))
	if err != nil {
		return err
	}

	c.client = k8sClient

	return nil
}

func loadConfig(cfgJSON *extapi.JSON) (hetznerDNSProviderConfig, error) {
	cfg := hetznerDNSProviderConfig{}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}

func stringFromSecretData(secretData *map[string][]byte, key string) (string, error) {
	data, ok := (*secretData)[key]
	if !ok {
		return "", fmt.Errorf("key %q not found in secret data", key)
	}
	return string(data), nil
}

func addTxtRecord(config internal.Config, ch *v1alpha1.ChallengeRequest) {
	slogger := zapLogger.Sugar()
	url := config.ApiUrl + "/records"

	name := recordName(ch.ResolvedFQDN, config.ZoneName)
	zoneId, err := searchZoneId(config)

	if err != nil {
		slogger.Errorf("unable to find id for zone name `%s`; %v", config.ZoneName, err)
	}

	var jsonStr = fmt.Sprintf(`{"value":"%s", "ttl":120, "type":"TXT", "name":"%s", "zone_id":"%s"}`, ch.Key, name, zoneId)

	add, err := callDnsApi(url, "POST", bytes.NewBuffer([]byte(jsonStr)), config)

	if err != nil {
		slogger.Error(err)
	}
	slogger.Infof("Added TXT record result: %s", string(add))
}

func clientConfig(c *hetznerDNSProviderSolver, ch *v1alpha1.ChallengeRequest) (internal.Config, error) {
	var config internal.Config

	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return config, err
	}
	config.ZoneName = cfg.ZoneName
	config.ApiUrl = cfg.ApiUrl
	secretName := cfg.SecretRef

	sec, err := c.client.CoreV1().Secrets(ch.ResourceNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})

	if err != nil {
		return config, fmt.Errorf("unable to get secret `%s/%s`; %v", secretName, ch.ResourceNamespace, err)
	}

	apiKey, err := stringFromSecretData(&sec.Data, "api-key")
	config.ApiKey = apiKey

	if err != nil {
		return config, fmt.Errorf("unable to get api-key from secret `%s/%s`; %v", secretName, ch.ResourceNamespace, err)
	}

	return config, nil
}

/*
Domain name in Hetzner is divided in 2 parts: record + zone name. API works
with record name that is FQDN without zone name. Sub-domains is a part of
record name and is separated by "."
*/
func recordName(fqdn string, domain string) string {
	slogger := zapLogger.Sugar()
	slogger.Infof("starting process on fqdn: %s for domain: %s", fqdn, domain)
	r := regexp.MustCompile("(.+)\\." + domain)
	name := r.FindStringSubmatch(fqdn)
	slogger.Infof("found record name : %s", name)
	if len(name) != 2 {
		slogger.Errorf("splitting domain name %s failed!", fqdn)
		return ""
	}
	return name[1]
}

func callDnsApi(url string, method string, body io.Reader, config internal.Config) ([]byte, error) {
	slogger := zapLogger.Sugar()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to execute request %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-API-Token", config.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slogger.Fatal(err)
		}
	}()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		return respBody, nil
	}

	text := "Error calling API status:" + resp.Status + " url: " + url + " method: " + method
	slogger.Error(text)
	return nil, errors.New(text)
}

func searchZoneId(config internal.Config) (string, error) {
	slogger := zapLogger.Sugar()
	// removing the dot so we can find the zone name correctly
	url := config.ApiUrl + "/zones?name=" + strings.TrimSuffix(config.ZoneName, ".")

	// Get Zone configuration
	zoneRecords, err := callDnsApi(url, "GET", nil, config)

	if err != nil {
		return "", fmt.Errorf("unable to get zone info %v", err)
	}

	// Unmarshall response
	zones := internal.ZoneResponse{}
	readErr := json.Unmarshal(zoneRecords, &zones)

	if readErr != nil {
		return "", fmt.Errorf("unable to unmarshal response %v", readErr)
	}

	if zones.Meta.Pagination.TotalEntries != 1 {
		return "", fmt.Errorf("wrong number of zones in response %d must be exactly = 1", zones.Meta.Pagination.TotalEntries)
	}

	slogger.Infof(" found zone name: %s  with total records: %v", zones.Zones[0].Name, zones.Zones[0].RecordsCount)
	return zones.Zones[0].Id, nil
}
