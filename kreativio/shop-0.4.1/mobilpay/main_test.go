package mobilpay

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/ssh"
)

func init() {
	cb := &CB{}
	cb.ParseKeys()

}
func Test_getPrivateKey(t *testing.T) {
	data, _ := ioutil.ReadFile("sandbox.LK1F-GMV1-YWRD-7J6T-QD55private.key")
	key, _ := ssh.ParseRawPrivateKey(data)
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *rsa.PrivateKey
		wantErr bool
	}{
		{name: "test #1", args: args{data: data}, wantErr: false, want: key.(*rsa.PrivateKey)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPrivateKey(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPrivateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPublicKey(t *testing.T) {
	data, _ := ioutil.ReadFile("sandbox.LK1F-GMV1-YWRD-7J6T-QD55.public.cer")
	b, _ := pem.Decode(data)
	cert, _ := x509.ParseCertificate(b.Bytes)
	pub := cert.PublicKey.(*rsa.PublicKey)
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *rsa.PublicKey
		wantErr bool
	}{
		{name: "test #1", args: args{data: data}, want: pub},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPublicKey(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_xmlMarshal(t *testing.T) {
	type x struct {
		Name string `xml:"name"`
	}
	w, _ := xml.Marshal(&x{})
	type args struct {
		rsp interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{name: "test #1", args: args{rsp: []byte("")}, wantErr: true},
		{name: "test #1", args: args{&x{}}, wantErr: false, want: w},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := xmlMarshal(tt.args.rsp)
			if (err != nil) != tt.wantErr {
				t.Errorf("xmlMarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("xmlMarshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncrypt(t *testing.T) {
	data, _ := ioutil.ReadFile("sandbox.LK1F-GMV1-YWRD-7J6T-QD55.public.cer")
	b, _ := pem.Decode(data)
	cert, _ := x509.ParseCertificate(b.Bytes)
	pub := cert.PublicKey.(*rsa.PublicKey)
	type args struct {
		publicKey  *rsa.PublicKey
		sourceText []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test #1", args: args{pub, []byte("text")}, wantErr: false},
		{name: "test #1", args: args{&rsa.PublicKey{}, []byte("text")}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := Encrypt(tt.args.publicKey, tt.args.sourceText)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	encTxt, encKey, _ := Encrypt(PrivateKey.Public().(*rsa.PublicKey), []byte("some text"))
	type args struct {
		encKey        string
		encryptedText string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test #1", args: args{"", ""}, wantErr: true},
		{name: "test #1", args: args{encKey, encTxt}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := Decrypt(tt.args.encKey, tt.args.encryptedText)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type FakeContextExecutor struct {
	db boil.Executor
}
type r struct {
}

func (r *r) LastInsertId() (int64, error) {
	return 1, nil
}
func (r *r) RowsAffected() (int64, error) {
	return 1, nil
}
func (f *FakeContextExecutor) Exec(query string, args ...interface{}) (sql.Result, error) {
	return &r{}, nil
}
func (f *FakeContextExecutor) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (f *FakeContextExecutor) QueryRow(query string, args ...interface{}) *sql.Row {
	return nil
}
func (f *FakeContextExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return &r{}, nil
}
func (f *FakeContextExecutor) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (f *FakeContextExecutor) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}
func prepRq(t *testing.T, b []byte) *http.Request {
	encTxt, encKey, e := Encrypt(PrivateKey.Public().(*rsa.PublicKey), b)
	if e != nil {
		t.Fail()
	}
	v := url.Values{}
	v.Add("env_key", encKey)
	v.Add("data", encTxt)
	rq, _ := http.NewRequest("POST", "", nil)
	rq.Form = v
	return rq
}
func TestCB_MobilpayConfirm(t *testing.T) {
	xmlrConfirmed := MResponse{}
	xmlrConfirmed.Order.Mobilpay.Action = "confirmed"
	confirmed, _ := xml.Marshal(&xmlrConfirmed)
	xmlrCpending := MResponse{}
	xmlrCpending.Order.Mobilpay.Action = "confirmed_pending"
	cpending, _ := xml.Marshal(&xmlrCpending)
	xmlrPpending := MResponse{}
	xmlrPpending.Order.Mobilpay.Action = "paid_pending"
	ppending, _ := xml.Marshal(&xmlrPpending)
	xmlrPaid := MResponse{}
	xmlrPaid.Order.Mobilpay.Action = "paid"
	paid, _ := xml.Marshal(&xmlrPaid)
	xmlrCanceled := MResponse{}
	xmlrCanceled.Order.Mobilpay.Action = "canceled"
	canceled, _ := xml.Marshal(&xmlrCanceled)
	xmlrCredit := MResponse{}
	xmlrCredit.Order.Mobilpay.Action = "credit"
	credit, _ := xml.Marshal(&xmlrCredit)
	xmlrDefault := MResponse{}
	xmlrDefault.Order.Mobilpay.Action = "foo"
	defaultt, _ := xml.Marshal(&xmlrDefault)

	type args struct {
		r  *http.Request
		wr http.ResponseWriter
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "test #1", args: args{prepRq(t, defaultt), httptest.NewRecorder()}},
		{name: "test #2", args: args{prepRq(t, confirmed), httptest.NewRecorder()}},
		{name: "test #3", args: args{prepRq(t, cpending), httptest.NewRecorder()}},
		{name: "test #4", args: args{prepRq(t, ppending), httptest.NewRecorder()}},
		{name: "test #5", args: args{prepRq(t, paid), httptest.NewRecorder()}},
		{name: "test #6", args: args{prepRq(t, canceled), httptest.NewRecorder()}},
		{name: "test #7", args: args{prepRq(t, credit), httptest.NewRecorder()}},
		{name: "test #8", args: args{new(http.Request), httptest.NewRecorder()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &CB{
				DBh: nil,
			}
			o.MobilpayConfirm(tt.args.wr, tt.args.r)
		})
	}
}

func TestSetMobilpayVars(t *testing.T) {
	type args struct {
		mobilpayEndpoint string
		signature        string
		privateKeyFile   string
		publicKeyFile    string
		confirmURL       string
		returnURL        string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "just call"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetMobilpayVars(tt.args.mobilpayEndpoint, tt.args.signature, tt.args.privateKeyFile, tt.args.publicKeyFile, tt.args.confirmURL, tt.args.returnURL)
		})
	}
}
