package home

import (
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"
	"howett.net/plist"
)

type DNSSettings struct {
	DNSProtocol string
	ServerURL   string `plist:",omitempty"`
	ServerName  string `plist:",omitempty"`
}

type PayloadContent = struct {
	Name               string
	PayloadDescription string
	PayloadDisplayName string
	PayloadIdentifier  string
	PayloadType        string
	PayloadUUID        string
	PayloadVersion     int
	DNSSettings        DNSSettings
}

type MobileConfig = struct {
	PayloadContent           []PayloadContent
	PayloadDescription       string
	PayloadDisplayName       string
	PayloadIdentifier        string
	PayloadRemovalDisallowed bool
	PayloadType              string
	PayloadUUID              string
	PayloadVersion           int
}

func genUUIDv4() string {
	return uuid.NewV4().String()
}

func getMobileConfig(w http.ResponseWriter, r *http.Request, d DNSSettings) []byte {
	data := MobileConfig{
		PayloadContent: []PayloadContent{{
			Name:               fmt.Sprintf("%s DNS over %s", r.Host, d.DNSProtocol),
			PayloadDescription: "Configures device to use AdGuard Home",
			PayloadDisplayName: "AdGuard Home",
			PayloadIdentifier:  fmt.Sprintf("com.apple.dnsSettings.managed.%s", genUUIDv4()),
			PayloadType:        "com.apple.dnsSettings.managed",
			PayloadUUID:        genUUIDv4(),
			PayloadVersion:     1,
			DNSSettings:        d,
		}},
		PayloadDescription:       "Adds AdGuard Home to Big Sur and iOS 14 or newer systems",
		PayloadDisplayName:       "AdGuard Home",
		PayloadIdentifier:        genUUIDv4(),
		PayloadRemovalDisallowed: false,
		PayloadType:              "Configuration",
		PayloadUUID:              genUUIDv4(),
		PayloadVersion:           1,
	}

	mobileconfig, err := plist.MarshalIndent(data, plist.XMLFormat, "\t")

	if err != nil {
		httpError(w, http.StatusInternalServerError, "plist.MarshalIndent: %s", err)
	}

	return mobileconfig
}

func handleMobileConfig(w http.ResponseWriter, r *http.Request, d DNSSettings) {
	mobileconfig := getMobileConfig(w, r, d)

	w.Header().Set("Content-Type", "application/xml")
	w.Write(mobileconfig)
}

func handleMobileConfigDoh(w http.ResponseWriter, r *http.Request) {
	handleMobileConfig(w, r, DNSSettings{
		DNSProtocol: "HTTPS",
		ServerURL:   fmt.Sprintf("https://%s/dns-query", r.Host),
	})
}

func handleMobileConfigDot(w http.ResponseWriter, r *http.Request) {
	handleMobileConfig(w, r, DNSSettings{
		DNSProtocol: "TLS",
		ServerName:  r.Host,
	})
}
