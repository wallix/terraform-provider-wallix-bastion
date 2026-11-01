package bastion

import (
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// OpenSSL short names (OBJ_nid2sn) the Bastion prints for subject attributes; any other OID prints as UNDEF.
var certificateSubjectAttributeNames = map[string]string{ //nolint:gochecknoglobals
	"0.9.2342.19200300.100.1.1":  "UID",
	"0.9.2342.19200300.100.1.25": "DC",
	"1.2.840.113549.1.9.1":       "emailAddress",
	"1.2.840.113549.1.9.2":       "unstructuredName",
	"1.3.6.1.4.1.311.60.2.1.1":   "jurisdictionL",
	"1.3.6.1.4.1.311.60.2.1.2":   "jurisdictionST",
	"1.3.6.1.4.1.311.60.2.1.3":   "jurisdictionC",
	"2.5.4.3":                    "CN",
	"2.5.4.4":                    "SN",
	"2.5.4.5":                    "serialNumber",
	"2.5.4.6":                    "C",
	"2.5.4.7":                    "L",
	"2.5.4.8":                    "ST",
	"2.5.4.9":                    "street",
	"2.5.4.10":                   "O",
	"2.5.4.11":                   "OU",
	"2.5.4.12":                   "title",
	"2.5.4.13":                   "description",
	"2.5.4.15":                   "businessCategory",
	"2.5.4.17":                   "postalCode",
	"2.5.4.18":                   "postOfficeBox",
	"2.5.4.20":                   "telephoneNumber",
	"2.5.4.41":                   "name",
	"2.5.4.42":                   "GN",
	"2.5.4.43":                   "initials",
	"2.5.4.44":                   "generationQualifier",
	"2.5.4.45":                   "x500UniqueIdentifier",
	"2.5.4.46":                   "dnQualifier",
	"2.5.4.65":                   "pseudonym",
	"2.5.4.72":                   "role",
	"2.5.4.97":                   "organizationIdentifier",
}

// Only the fields up to the subject: x509.ParseCertificate rejects subjects that OpenSSL loads.
type certificateUpToSubject struct {
	TBSCertificate struct {
		Version      int `asn1:"optional,explicit,default:0,tag:0"`
		SerialNumber asn1.RawValue
		Signature    asn1.RawValue
		Issuer       asn1.RawValue
		Validity     asn1.RawValue
		Subject      asn1.RawValue
	}
}

type certificateSubjectAttribute struct {
	Type  asn1.ObjectIdentifier
	Value asn1.RawValue
}

// The SET suffix makes encoding/asn1 read a SET OF.
type certificateSubjectRDNSET []certificateSubjectAttribute

// certificateSubject returns the subject of the first certificate in pemData the way the Bastion API
// reads back a certificate: "/<short name>=<value>" per attribute in DER order, raw value, no escaping.
func certificateSubject(pemData string) (string, error) {
	rest := []byte(pemData)
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			return "", errors.New("no PEM certificate found")
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		var cert certificateUpToSubject
		if _, err := asn1.Unmarshal(block.Bytes, &cert); err != nil {
			return "", fmt.Errorf("parsing certificate: %w", err)
		}

		return formatCertificateSubject(cert.TBSCertificate.Subject.FullBytes)
	}
}

func formatCertificateSubject(rawSubject []byte) (string, error) {
	var rdns []certificateSubjectRDNSET
	rest, err := asn1.Unmarshal(rawSubject, &rdns)
	if err != nil {
		return "", fmt.Errorf("parsing certificate subject: %w", err)
	}
	if len(rest) > 0 {
		return "", errors.New("parsing certificate subject: trailing data")
	}
	var subject strings.Builder
	for _, rdn := range rdns {
		for _, attr := range rdn {
			// The API decodes the raw value as UTF-8 (HTTP 500 on T61String or BMPString with non-ASCII).
			if !utf8.Valid(attr.Value.Bytes) {
				return "", fmt.Errorf("certificate subject attribute %s is not valid UTF-8", attr.Type)
			}
			name, ok := certificateSubjectAttributeNames[attr.Type.String()]
			if !ok {
				name = "UNDEF"
			}
			subject.WriteString("/" + name + "=")
			subject.Write(attr.Value.Bytes)
		}
	}

	return subject.String(), nil
}

// certificateStateValue returns the value to store for a certificate the API reads back as its subject:
// the PEM from the prior state while its subject matches the API, the API value otherwise.
func certificateStateValue(priorPEM, apiSubject string) string {
	if apiSubject == "" {
		return ""
	}
	if subject, err := certificateSubject(priorPEM); err == nil && subject == apiSubject {
		return priorPEM
	}

	return apiSubject
}
