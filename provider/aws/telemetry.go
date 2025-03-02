package aws

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	skipParams = strings.Join([]string{
		"DefaultAmi",
		"DefaultAmiArm",
		"Password",
		"Telemetry",
		"Version",
	}, ",")

	redactedParams = strings.Join([]string{
		"Ami",
		"ApiCount",
		"ApiCpu",
		"ApiMonitorMemory",
		"BuildInstancePolicy",
		"ClientId",
		"ExistingVpc",
		"HttpProxy",
		"InternalSuffix",
		"InstanceBootCommand",
		"InstanceRunCommand",
		"InstancePolicy",
		"InstanceSecurityGroup",
		"BuildInstanceSecurityGroup",
		"InternetGateway",
		"Key",
		"LogBucket",
		"RouterInternalSecurityGroup",
		"RouterSecurityGroup",
		"SslPolicy",
		"SyslogDestination",
		"WhiteList",
	}, ",")
)

func (p *Provider) RackParamsToSync(release string, params map[string]string) map[string]string {
	defaultParams, err := templateParams(release)
	if err != nil {
		return nil
	}

	toSync := make(map[string]string)
	for k, v := range params {
		if strings.Contains(skipParams, k) {
			continue
		}

		if v2, ok := defaultParams[k]; ok && v == v2 {
			continue
		}

		if strings.Contains(redactedParams, k) {
			v = hashParamValue(v)
		}

		toSync[k] = v
	}

	return toSync
}

func templateParams(release string) (map[string]string, error) {
	var template []byte

	if release == "dev" {
		data, err := os.ReadFile("provider/aws/formation/rack.json")
		if err != nil {
			return nil, err
		}

		template = data
	} else {
		res, err := http.Get(fmt.Sprintf("https://convox.s3.amazonaws.com/release/%s/rack.json", release))
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		template = data
	}

	var templateMap map[string]interface{}
	if err := json.Unmarshal([]byte(template), &templateMap); err != nil {
		return nil, err
	}

	params := map[string]string{}
	for k, v := range templateMap["Parameters"].(map[string]interface{}) {
		d, ok := v.(map[string]interface{})["Default"]
		if ok {
			params[k] = d.(string)
		}
	}

	return params, nil
}

func hashParamValue(value string) string {
	hasher := sha256.New()
	hasher.Write([]byte(value))
	return hex.EncodeToString(hasher.Sum(nil))
}
