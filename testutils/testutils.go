/*
Sniperkit-Bot
- Status: analyzed
*/

package testutils

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"

	routelib "github.com/vulcand/route"
	"golang.org/x/crypto/ocsp"

	"github.com/sniperkit/snk.fork.vulcand/engine"
	"github.com/sniperkit/snk.fork.vulcand/plugin/ratelimit"
)

func init() {
	bytes, err := hex.DecodeString(OCSPResponseHex)
	if err != nil {
		panic(err)
	}
	OCSPResponseBytes = bytes
	OCSPResponse, err = ocsp.ParseResponse(bytes, nil)
	if err != nil {
		panic(err)
	}
}

var lastId int64

func UID(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, atomic.AddInt64(&lastId, 1))
}

type Batch struct {
	Route    string
	Addr     string
	URL      string
	Protocol string
	Host     string
	KeyPair  *engine.KeyPair
	AutoCert *engine.AutoCertSettings
}

type BatchVal struct {
	H engine.Host

	L  engine.Listener
	LK engine.ListenerKey

	F  engine.Frontend
	FK engine.FrontendKey

	B  engine.Backend
	BK engine.BackendKey

	S  engine.Server
	SK engine.ServerKey
}

func MakeURL(l engine.Listener, path string) string {
	return fmt.Sprintf("%s://%s%s", l.Protocol, l.Address.Address, path)
}

func (b BatchVal) FrontendURL(path string) string {
	return MakeURL(b.L, path)
}

func MakeBatch(b Batch) (bv BatchVal) {
	if b.Host == "" {
		b.Host = "localhost"
	}
	be := MakeBackend()
	beSrv := MakeServer(b.URL)

	bv.H = MakeHost(b.Host, b.KeyPair, b.AutoCert)
	if b.Addr != "" {
		if b.Protocol == "" {
			b.Protocol = engine.HTTP
		}
		bv.L = MakeListener(b.Addr, b.Protocol)
		bv.LK = engine.ListenerKey{Id: bv.L.Id}
	}
	bv.F = MakeFrontend(b.Route, be.Id)
	bv.FK = engine.FrontendKey{Id: bv.F.Id}
	bv.B = be
	bv.BK = engine.BackendKey{Id: be.Id}
	bv.S = beSrv
	bv.SK = engine.ServerKey{BackendKey: engine.BackendKey{Id: be.Id}, Id: beSrv.Id}
	return bv
}

func (bv *BatchVal) Snapshot() engine.Snapshot {
	return engine.Snapshot{
		Hosts:     []engine.Host{bv.H},
		Listeners: []engine.Listener{bv.L},
		BackendSpecs: []engine.BackendSpec{
			{Backend: bv.B, Servers: []engine.Server{bv.S}},
		},
		FrontendSpecs: []engine.FrontendSpec{
			{Frontend: bv.F},
		},
	}
}

func MakeSnapshot(bvs ...BatchVal) engine.Snapshot {
	var ss engine.Snapshot
	for _, bv := range bvs {
		bss := bv.Snapshot()
		ss.Hosts = append(ss.Hosts, bss.Hosts...)
		ss.BackendSpecs = append(ss.BackendSpecs, bss.BackendSpecs...)
		ss.Listeners = append(ss.Listeners, bss.Listeners...)
		ss.FrontendSpecs = append(ss.FrontendSpecs, bss.FrontendSpecs...)
	}
	return ss
}

func MakeHost(name string, keyPair *engine.KeyPair, autoCert *engine.AutoCertSettings) engine.Host {
	return engine.Host{
		Name:     name,
		Settings: engine.HostSettings{KeyPair: keyPair, AutoCert: autoCert},
	}
}

func MakeListener(addr string, protocol string) engine.Listener {
	l, err := engine.NewListener(fmt.Sprintf("listener_%v", addr), protocol, engine.TCP, addr, "", "", nil)
	if err != nil {
		panic(err)
	}
	return *l
}

func MakeFrontend(route string, backendId string) engine.Frontend {
	f, err := engine.NewHTTPFrontend(routelib.NewMux(), UID("frontend"), backendId, route, engine.HTTPFrontendSettings{})
	if err != nil {
		panic(err)
	}
	return *f
}

func MakeBackend() engine.Backend {
	b, err := engine.NewHTTPBackend(UID("backend"), engine.HTTPBackendSettings{})
	if err != nil {
		panic(err)
	}
	return *b
}

func MakeServer(url string) engine.Server {
	s, err := engine.NewServer(UID("server"), url)
	if err != nil {
		panic(err)
	}
	return *s
}

func MakeRateLimit(id string, rate int64, variable string, burst int64, periodSeconds int64) engine.Middleware {
	rl, err := ratelimit.FromOther(ratelimit.RateLimit{
		PeriodSeconds: periodSeconds,
		Requests:      rate,
		Burst:         burst,
		Variable:      variable})
	if err != nil {
		panic(err)
	}
	return engine.Middleware{
		Type:       "ratelimit",
		Id:         id,
		Middleware: rl,
	}
}

func NewTestKeyPair() *engine.KeyPair {
	return &engine.KeyPair{
		Key:  LocalhostKey,
		Cert: LocalhostCert,
	}
}

func NewOCSPResponder() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "ocsp-response")
		w.Write(OCSPResponseBytes)
	}))
}

var LocalhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIIBWjCCAQSgAwIBAgIJANX7GqdOyHEWMA0GCSqGSIb3DQEBCwUAMBIxEDAOBgNV
BAMMB3Rlc3QtY2EwHhcNMTcwOTI3MjMwNzQ4WhcNMjcwOTI1MjMwNzQ4WjAWMRQw
EgYDVQQDDAtleGFtcGxlLmNvbTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQC5dZMr
DRdtCehGqg9FdXtMrtdGouawgcun+Aq9qd02PzNTwpGhPupeD3UGP2s31b8Gq+B9
Jutk1Ra8W7rfAw+7AgMBAAGjOTA3MAkGA1UdEwQCMAAwCwYDVR0PBAQDAgXgMB0G
A1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcDATANBgkqhkiG9w0BAQsFAANBAIa1
mpM7OAqeLRLtYDlti4Ydop6OTMlAV9lrOww5N2XqD5A0h7tlyz+sxSHmxLdORpQT
ntb2OVkmSvA6UXZwerg=
-----END CERTIFICATE-----`)

// localhostKey is the private key for localhostCert.
var LocalhostKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBALl1kysNF20J6EaqD0V1e0yu10ai5rCBy6f4Cr2p3TY/M1PCkaE+
6l4PdQY/azfVvwar4H0m62TVFrxbut8DD7sCAwEAAQJAS+y0eTV816jsrLFGWztD
ZRtXCpX6N1yL0ZIcY5U8+M2oRgnI8TGUU3Qnibgc9MneDq9FCGpIPWGZvoTGQyn6
IQIhAOUQ/2ThTIwkogkdKkR3tOXiX5wfnmNqFn2vQ3wODwPTAiEAz0P7sIwrVfOU
IbdugtspRl5HYmydktdKJFZx/F6RG3kCIAl6opbyG9DQ00O9STp8GahJrdswE8XZ
ZgTdc9V2X3ixAiAeYtEbaPFRgIxPBR1mgcrjTh8ZBuNzC60K9sFvRk3vwQIhANNq
aGa1p0gO3OWxab9IlXoUYQN9kiQD3vOy/zTnx0As
-----END RSA PRIVATE KEY-----`)

var LocalhostCertChain = []byte(`-----BEGIN CERTIFICATE-----
MIIBWjCCAQSgAwIBAgIJANX7GqdOyHEWMA0GCSqGSIb3DQEBCwUAMBIxEDAOBgNV
BAMMB3Rlc3QtY2EwHhcNMTcwOTI3MjMwNzQ4WhcNMjcwOTI1MjMwNzQ4WjAWMRQw
EgYDVQQDDAtleGFtcGxlLmNvbTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQC5dZMr
DRdtCehGqg9FdXtMrtdGouawgcun+Aq9qd02PzNTwpGhPupeD3UGP2s31b8Gq+B9
Jutk1Ra8W7rfAw+7AgMBAAGjOTA3MAkGA1UdEwQCMAAwCwYDVR0PBAQDAgXgMB0G
A1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcDATANBgkqhkiG9w0BAQsFAANBAIa1
mpM7OAqeLRLtYDlti4Ydop6OTMlAV9lrOww5N2XqD5A0h7tlyz+sxSHmxLdORpQT
ntb2OVkmSvA6UXZwerg=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIBWjCCAQSgAwIBAgIJANX7GqdOyHEWMA0GCSqGSIb3DQEBCwUAMBIxEDAOBgNV
BAMMB3Rlc3QtY2EwHhcNMTcwOTI3MjMwNzQ4WhcNMjcwOTI1MjMwNzQ4WjAWMRQw
EgYDVQQDDAtleGFtcGxlLmNvbTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQC5dZMr
DRdtCehGqg9FdXtMrtdGouawgcun+Aq9qd02PzNTwpGhPupeD3UGP2s31b8Gq+B9
Jutk1Ra8W7rfAw+7AgMBAAGjOTA3MAkGA1UdEwQCMAAwCwYDVR0PBAQDAgXgMB0G
A1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcDATANBgkqhkiG9w0BAQsFAANBAIa1
mpM7OAqeLRLtYDlti4Ydop6OTMlAV9lrOww5N2XqD5A0h7tlyz+sxSHmxLdORpQT
ntb2OVkmSvA6UXZwerg=
-----END CERTIFICATE-----`)

// Took from golang.org/x/crypto/ocsp
const OCSPResponseHex = "308206bc0a0100a08206b5308206b106092b0601050507300101048206a23082069e3081" +
	"c9a14e304c310b300906035504061302494c31163014060355040a130d5374617274436f" +
	"6d204c74642e312530230603550403131c5374617274436f6d20436c6173732031204f43" +
	"5350205369676e6572180f32303130303730373137333531375a30663064303c30090605" +
	"2b0e03021a050004146568874f40750f016a3475625e1f5c93e5a26d580414eb4234d098" +
	"b0ab9ff41b6b08f7cc642eef0e2c45020301d0fa8000180f323031303037303731353031" +
	"30355aa011180f32303130303730373138333531375a300d06092a864886f70d01010505" +
	"000382010100ab557ff070d1d7cebbb5f0ec91a15c3fed22eb2e1b8244f1b84545f013a4" +
	"fb46214c5e3fbfbebb8a56acc2b9db19f68fd3c3201046b3824d5ba689f99864328710cb" +
	"467195eb37d84f539e49f859316b32964dc3e47e36814ce94d6c56dd02733b1d0802f7ff" +
	"4eebdbbd2927dcf580f16cbc290f91e81b53cb365e7223f1d6e20a88ea064104875e0145" +
	"672b20fc14829d51ca122f5f5d77d3ad6c83889c55c7dc43680ba2fe3cef8b05dbcabdc0" +
	"d3e09aaf9725597f8c858c2fa38c0d6aed2e6318194420dd1a1137445d13e1c97ab47896" +
	"17a4e08925f46f867b72e3a4dc1f08cb870b2b0717f7207faa0ac512e628a029aba7457a" +
	"e63dcf3281e2162d9349a08204ba308204b6308204b23082039aa003020102020101300d" +
	"06092a864886f70d010105050030818c310b300906035504061302494c31163014060355" +
	"040a130d5374617274436f6d204c74642e312b3029060355040b13225365637572652044" +
	"69676974616c204365727469666963617465205369676e696e6731383036060355040313" +
	"2f5374617274436f6d20436c6173732031205072696d61727920496e7465726d65646961" +
	"746520536572766572204341301e170d3037313032353030323330365a170d3132313032" +
	"333030323330365a304c310b300906035504061302494c31163014060355040a130d5374" +
	"617274436f6d204c74642e312530230603550403131c5374617274436f6d20436c617373" +
	"2031204f435350205369676e657230820122300d06092a864886f70d0101010500038201" +
	"0f003082010a0282010100b9561b4c45318717178084e96e178df2255e18ed8d8ecc7c2b" +
	"7b51a6c1c2e6bf0aa3603066f132fe10ae97b50e99fa24b83fc53dd2777496387d14e1c3" +
	"a9b6a4933e2ac12413d085570a95b8147414a0bc007c7bcf222446ef7f1a156d7ea1c577" +
	"fc5f0facdfd42eb0f5974990cb2f5cefebceef4d1bdc7ae5c1075c5a99a93171f2b0845b" +
	"4ff0864e973fcfe32f9d7511ff87a3e943410c90a4493a306b6944359340a9ca96f02b66" +
	"ce67f028df2980a6aaee8d5d5d452b8b0eb93f923cc1e23fcccbdbe7ffcb114d08fa7a6a" +
	"3c404f825d1a0e715935cf623a8c7b59670014ed0622f6089a9447a7a19010f7fe58f841" +
	"29a2765ea367824d1c3bb2fda308530203010001a382015c30820158300c0603551d1301" +
	"01ff04023000300b0603551d0f0404030203a8301e0603551d250417301506082b060105" +
	"0507030906092b0601050507300105301d0603551d0e0416041445e0a36695414c5dd449" +
	"bc00e33cdcdbd2343e173081a80603551d230481a030819d8014eb4234d098b0ab9ff41b" +
	"6b08f7cc642eef0e2c45a18181a47f307d310b300906035504061302494c311630140603" +
	"55040a130d5374617274436f6d204c74642e312b3029060355040b132253656375726520" +
	"4469676974616c204365727469666963617465205369676e696e67312930270603550403" +
	"13205374617274436f6d2043657274696669636174696f6e20417574686f726974798201" +
	"0a30230603551d12041c301a8618687474703a2f2f7777772e737461727473736c2e636f" +
	"6d2f302c06096086480186f842010d041f161d5374617274436f6d205265766f63617469" +
	"6f6e20417574686f72697479300d06092a864886f70d01010505000382010100182d2215" +
	"8f0fc0291324fa8574c49bb8ff2835085adcbf7b7fc4191c397ab6951328253fffe1e5ec" +
	"2a7da0d50fca1a404e6968481366939e666c0a6209073eca57973e2fefa9ed1718e8176f" +
	"1d85527ff522c08db702e3b2b180f1cbff05d98128252cf0f450f7dd2772f4188047f19d" +
	"c85317366f94bc52d60f453a550af58e308aaab00ced33040b62bf37f5b1ab2a4f7f0f80" +
	"f763bf4d707bc8841d7ad9385ee2a4244469260b6f2bf085977af9074796048ecc2f9d48" +
	"a1d24ce16e41a9941568fec5b42771e118f16c106a54ccc339a4b02166445a167902e75e" +
	"6d8620b0825dcd18a069b90fd851d10fa8effd409deec02860d26d8d833f304b10669b42"

var OCSPResponse *ocsp.Response
var OCSPResponseBytes []byte
