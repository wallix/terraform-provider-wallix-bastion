package bastion

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	testOIDCN     = "2.5.4.3"
	testCASubject = "/CN=tfagent195 ca"
	// "tfagent195 bmp" as BMPString content (UCS-2).
	testBMPASCII = "\x00t\x00f\x00a\x00g\x00e\x00n\x00t\x001\x009\x005\x00 \x00b\x00m\x00p"
)

func testSubjectOID(t *testing.T, dotted string) asn1.ObjectIdentifier {
	t.Helper()
	arcs := strings.Split(dotted, ".")
	oid := make(asn1.ObjectIdentifier, 0, len(arcs))
	for _, arc := range arcs {
		n, err := strconv.Atoi(arc)
		if err != nil {
			t.Fatalf("bad OID %q: %v", dotted, err)
		}
		oid = append(oid, n)
	}

	return oid
}

// testSubjectRDN is one RDN: OID, ASN.1 string tag and raw value for each attribute.
type testSubjectRDN []struct {
	oid   string
	tag   int
	value string
}

func testUTF8RDNs(pairs ...string) []testSubjectRDN {
	rdns := make([]testSubjectRDN, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		rdns = append(rdns, testSubjectRDN{{pairs[i], asn1.TagUTF8String, pairs[i+1]}})
	}

	return rdns
}

func testCertificatePEM(t *testing.T, rdns []testSubjectRDN) string {
	t.Helper()
	sets := make([]certificateSubjectRDNSET, 0, len(rdns))
	for _, rdn := range rdns {
		set := make(certificateSubjectRDNSET, 0, len(rdn))
		for _, a := range rdn {
			set = append(set, certificateSubjectAttribute{
				Type:  testSubjectOID(t, a.oid),
				Value: asn1.RawValue{Tag: a.tag, Bytes: []byte(a.value)},
			})
		}
		sets = append(sets, set)
	}
	rawSubject, err := asn1.Marshal(sets)
	if err != nil {
		t.Fatal(err)
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		RawSubject:            rawSubject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

// Expected values are what GET /externalauths returned for ca_certificate on a Bastion 12.4.3
// (API v3.12 and v3.8, identical), for CAs with the same subject made with openssl.
func TestCertificateSubject(t *testing.T) {
	t.Parallel()
	ou := make([]string, 0, 22)
	ou = append(ou, testOIDCN, "tfagent195 long")
	var long strings.Builder
	long.WriteString("/CN=tfagent195 long")
	for i := 1; i <= 10; i++ {
		v := "ou" + strconv.Itoa(i) + "-" + strings.Repeat("a", 55)
		ou = append(ou, "2.5.4.11", v)
		long.WriteString("/OU=" + v)
	}
	cases := []struct {
		name    string
		rdns    []testSubjectRDN
		want    string
		wantErr bool
	}{
		{name: "simple", rdns: testUTF8RDNs(testOIDCN, "tfagent195 simple"), want: "/CN=tfagent195 simple"},
		{
			name: "order",
			rdns: testUTF8RDNs("2.5.4.6", "FR", "2.5.4.8", "Ile-de-France", "2.5.4.7", "Paris", "2.5.4.10", "WALLIX",
				"2.5.4.11", "R&D", testOIDCN, "tfagent195 order1", "1.2.840.113549.1.9.1", "ca@example.com"),
			want: "/C=FR/ST=Ile-de-France/L=Paris/O=WALLIX/OU=R&D/CN=tfagent195 order1/emailAddress=ca@example.com",
		},
		{
			name: "reverse order",
			rdns: testUTF8RDNs("1.2.840.113549.1.9.1", "ca@example.com", testOIDCN, "tfagent195 order2",
				"2.5.4.11", "R&D", "2.5.4.10", "WALLIX", "2.5.4.7", "Paris", "2.5.4.8", "IDF", "2.5.4.6", "FR"),
			want: "/emailAddress=ca@example.com/CN=tfagent195 order2/OU=R&D/O=WALLIX/L=Paris/ST=IDF/C=FR",
		},
		{
			name: "domain components",
			rdns: testUTF8RDNs("0.9.2342.19200300.100.1.25", "com", "0.9.2342.19200300.100.1.25", "example",
				testOIDCN, "tfagent195 dc"),
			want: "/DC=com/DC=example/CN=tfagent195 dc",
		},
		{
			// openssl -subj "/O=WALLIX/OU=Eng+CN=tfagent195 multi1": no "+", attributes in DER order.
			name: "multi-valued RDN",
			rdns: []testSubjectRDN{
				{{"2.5.4.10", asn1.TagUTF8String, "WALLIX"}},
				{{"2.5.4.11", asn1.TagUTF8String, "Eng"}, {testOIDCN, asn1.TagUTF8String, "tfagent195 multi1"}},
			},
			want: "/O=WALLIX/OU=Eng/CN=tfagent195 multi1",
		},
		{
			// openssl -subj "/CN=tfagent195 multi2+UID=u1+O=Org/C=FR", DER sorts the SET.
			name: "multi-valued RDN in DER order",
			rdns: []testSubjectRDN{
				{
					{"2.5.4.10", asn1.TagUTF8String, "Org"},
					{"0.9.2342.19200300.100.1.1", asn1.TagUTF8String, "u1"},
					{testOIDCN, asn1.TagUTF8String, "tfagent195 multi2"},
				},
				{{"2.5.4.6", asn1.TagUTF8String, "FR"}},
			},
			want: "/O=Org/UID=u1/CN=tfagent195 multi2/C=FR",
		},
		{name: "slash", rdns: testUTF8RDNs(testOIDCN, "tfagent195 a/b"), want: "/CN=tfagent195 a/b"},
		{name: "equal", rdns: testUTF8RDNs(testOIDCN, "tfagent195 x=y"), want: "/CN=tfagent195 x=y"},
		{name: "comma", rdns: testUTF8RDNs(testOIDCN, "tfagent195 a,b, c"), want: "/CN=tfagent195 a,b, c"},
		{name: "plus", rdns: testUTF8RDNs(testOIDCN, "tfagent195 p+q"), want: "/CN=tfagent195 p+q"},
		{name: "quotes", rdns: testUTF8RDNs(testOIDCN, `tfagent195 "dq" 'sq'`), want: `/CN=tfagent195 "dq" 'sq'`},
		{
			name: "spaces",
			rdns: testUTF8RDNs(testOIDCN, " tfagent195 spaces ", "2.5.4.10", "a  b"),
			want: "/CN= tfagent195 spaces /O=a  b",
		},
		{name: "backslash", rdns: testUTF8RDNs(testOIDCN, `tfagent195 back\slash`), want: `/CN=tfagent195 back\slash`},
		{
			name: "punctuation",
			rdns: testUTF8RDNs(testOIDCN, "tfagent195 <>;#~!*()&%$@"),
			want: "/CN=tfagent195 <>;#~!*()&%$@",
		},
		{
			name: "UTF-8",
			rdns: testUTF8RDNs(testOIDCN, "tfagent195 é", "2.5.4.10", "Société Générale", "2.5.4.7", "Zürich ✓"),
			want: "/CN=tfagent195 é/O=Société Générale/L=Zürich ✓",
		},
		{
			name: "attribute names",
			rdns: testUTF8RDNs("2.5.4.5", "123", "2.5.4.12", "T", "2.5.4.42", "Given", "2.5.4.4", "Sur",
				"2.5.4.9", "1 rue", "2.5.4.17", "75001", "0.9.2342.19200300.100.1.1", "uid1", "2.5.4.13", "D",
				"2.5.4.65", "P", "2.5.4.43", "I", "2.5.4.46", "Q", "2.5.4.44", "G", "2.5.4.41", "N",
				"2.5.4.15", "B", "2.5.4.97", "VATFR-1", "1.3.6.1.4.1.311.60.2.1.3", "FR", testOIDCN, "tfagent195 attrs"),
			want: "/serialNumber=123/title=T/GN=Given/SN=Sur/street=1 rue/postalCode=75001/UID=uid1/description=D" +
				"/pseudonym=P/initials=I/dnQualifier=Q/generationQualifier=G/name=N/businessCategory=B" +
				"/organizationIdentifier=VATFR-1/jurisdictionC=FR/CN=tfagent195 attrs",
		},
		{
			name: "more attribute names",
			rdns: testUTF8RDNs("2.5.4.72", "R", "1.2.840.113549.1.9.2", "U", "1.3.6.1.4.1.311.60.2.1.1", "JL",
				"1.3.6.1.4.1.311.60.2.1.2", "JST", "2.5.4.18", "42", "2.5.4.20", "0102", "2.5.4.45", "X",
				testOIDCN, "tfagent195 attrs2"),
			want: "/role=R/unstructuredName=U/jurisdictionL=JL/jurisdictionST=JST/postOfficeBox=42" +
				"/telephoneNumber=0102/x500UniqueIdentifier=X/CN=tfagent195 attrs2",
		},
		{
			name: "unknown OID",
			rdns: testUTF8RDNs("1.2.3.4", "tfagent195 oid", testOIDCN, "x"),
			want: "/UNDEF=tfagent195 oid/CN=x",
		},
		{name: "long subject", rdns: testUTF8RDNs(ou...), want: long.String()},
		{
			name: "PrintableString and T61String, ASCII",
			rdns: []testSubjectRDN{
				{{testOIDCN, asn1.TagT61String, "tfagent195 t61"}},
				{{"2.5.4.10", asn1.TagPrintableString, "plain ascii"}},
			},
			want: "/CN=tfagent195 t61/O=plain ascii",
		},
		{
			// Raw UCS-2 bytes, NUL included.
			name: "BMPString, ASCII",
			rdns: []testSubjectRDN{{{testOIDCN, asn1.TagBMPString, testBMPASCII}}},
			want: "/CN=" + testBMPASCII,
		},
		{
			// GET answers HTTP 500 for these two.
			name:    "T61String, non-ASCII",
			rdns:    []testSubjectRDN{{{testOIDCN, asn1.TagT61String, "tfagent195 t61 \xe9"}}},
			wantErr: true,
		},
		{
			name:    "BMPString, non-ASCII",
			rdns:    []testSubjectRDN{{{testOIDCN, asn1.TagBMPString, "\x00\xe9"}}},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := certificateSubject(testCertificatePEM(t, tc.rdns))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("got %q, want an error", got)
				}

				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got  %q\nwant %q", got, tc.want)
			}
		})
	}
}

func TestCertificateSubjectPEM(t *testing.T) {
	t.Parallel()
	first := testCertificatePEM(t, testUTF8RDNs(testOIDCN, "first"))
	second := testCertificatePEM(t, testUTF8RDNs(testOIDCN, "second"))
	key := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte{0x30, 0x00}}))
	cases := []struct {
		name    string
		pem     string
		want    string
		wantErr bool
	}{
		{name: "CRLF", pem: strings.ReplaceAll(first, "\n", "\r\n"), want: "/CN=first"},
		{name: "text before", pem: "Certificate:\n    Data: ...\n" + first, want: "/CN=first"},
		{name: "chain keeps the first certificate", pem: first + second, want: "/CN=first"},
		{name: "other block first", pem: key + second, want: "/CN=second"},
		{name: "empty", pem: "", wantErr: true},
		{name: "subject read back by the API", pem: "/CN=first", wantErr: true},
		{name: "not a certificate", pem: "-----BEGIN CERTIFICATE-----\nMIIBAAAA\n-----END CERTIFICATE-----\n", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := certificateSubject(tc.pem)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("got %q, want an error", got)
				}

				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// x509.ParseCertificate rejects a PrintableString with "_"; the Bastion takes it and returns /CN=tfagent195_ca.
func TestCertificateSubjectLenientStrings(t *testing.T) {
	t.Parallel()
	certPEM := testCertificatePEM(t, []testSubjectRDN{{{testOIDCN, asn1.TagPrintableString, "tfagent195_ca"}}})
	block, _ := pem.Decode([]byte(certPEM))
	if _, err := x509.ParseCertificate(block.Bytes); err == nil {
		t.Fatal("x509.ParseCertificate accepted the certificate, the test no longer covers the case")
	}
	got, err := certificateSubject(certPEM)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/CN=tfagent195_ca" {
		t.Errorf("got %q", got)
	}
}

func TestCertificateStateValue(t *testing.T) {
	t.Parallel()
	ca := testCertificatePEM(t, testUTF8RDNs(testOIDCN, "tfagent195 ca"))
	cases := []struct {
		name, prior, api, want string
	}{
		{name: "same subject keeps the PEM", prior: ca, api: testCASubject, want: ca},
		{name: "other subject", prior: ca, api: "/CN=other", want: "/CN=other"},
		{name: "removed outside of Terraform", prior: ca, api: "", want: ""},
		{name: "import", prior: "", api: testCASubject, want: testCASubject},
		{name: "0.17.0 state", prior: testCASubject, api: testCASubject, want: testCASubject},
		{name: "no CA", prior: "", api: "", want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := certificateStateValue(tc.prior, tc.api); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
