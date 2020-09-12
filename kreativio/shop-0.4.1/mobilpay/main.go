package mobilpay

import (
	"bytes"
	"crypto/rand"
	"crypto/rc4"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/moapis/multidb"
	"github.com/moapis/shop/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/ssh"
)

// MobilpayEndpoint - Mobilpay gateway
var MobilpayEndpoint = "http://sandboxsecure.mobilpay.ro"

// Signature - merchant signature
var Signature = "LK1F-GMV1-YWRD-7J6T-QD55"

// PrivateKeyFile - Private key of the mobilpay merchant, used to decript the confirm response.
var PrivateKeyFile = "sandbox.LK1F-GMV1-YWRD-7J6T-QD55private.key" // mobilpay test sandbox key

// PublicCer - Public key used to encript the requests made to mobilpay.
var PublicCer = "sandbox.LK1F-GMV1-YWRD-7J6T-QD55.public.cer" // mobilpay test sandbox cer
// ConfirmURL -
var ConfirmURL = ""

// ReturnURL -
var ReturnURL = ""

// SetMobilpayVars -
func SetMobilpayVars(mobilpayEndpoint, signature, privateKeyFile, publicKeyFile, confirmURL, returnURL string) {
	MobilpayEndpoint = mobilpayEndpoint
	Signature = signature
	PrivateKeyFile = privateKeyFile
	PublicCer = publicKeyFile
	ConfirmURL = confirmURL
	ReturnURL = returnURL
}

// ParseKeys - reads the files that PrivateKeyFile and PublicCer point to.
func (o *CB) ParseKeys() {
	err := make([]error, 4)
	var cer, b []byte
	cer, err[0] = ioutil.ReadFile(PublicCer)
	Cert, err[1] = getPublicKey(cer)
	b, err[2] = ioutil.ReadFile(PrivateKeyFile)
	PrivateKey, err[3] = getPrivateKey(b)
	for k, v := range err {
		if v != nil {
			log.Printf("CB init error #%d - %s", k, v.Error())
		}
	}
}

// CB - contains the http cb method and depends on DBh *multidb.MultiDB handle
type CB struct {
	DBh *multidb.MultiDB
}
type helper interface {
	xmlMarshal(rsp interface{}) []byte
}

// H - helper
type H struct{}

type crc struct {
	ErrorType int `xml:"error_type,attr"`
	ErrorCode int `xml:"error_code,attr"`
}

/*
MerchantsResponseE - represents the ping back that the confirm endpoint will make with attributes that specify error.
	ErrorType - 1 = temp error ; 2 = perm error ;
*/
type MerchantsResponseE struct {
	CRC crc `xml:"crc"`
}

// MerchantsResponse -
type MerchantsResponse struct {
	CRC string `xml:"crc"`
}

type customer struct {
	Type    string `xml:"type,attr"`
	Fname   string `xml:"first_name"`
	Lname   string `xml:"last_name"`
	Address string `xml:"address"`
	Email   string `xml:"email"`
	Phone   string `xml:"mobile_phone"`
}

type mobilpay struct {
	Timestamp           string   `xml:"timestamp,attr"`
	CRC                 string   `xml:"crc,attr"`
	Action              string   `xml:"action"`
	Customer            customer `xml:"customer"`
	Purchase            string   `xml:"purchase"`
	OriginalAmount      string   `xml:"original_amount"`
	ProcessedAmount     string   `xml:"processed_amount"`
	PanMasked           string   `xml:"pan_masked"`
	PaymentInstrumentID string   `xml:"payment_instrument_id"`
	TokenID             string   `xml:"token_id"`
	TokenExpirationDate string   `xml:"token_expiration_date"`
	ErrorCode           int      `xml:"error_code"`
}

// MResponse - is the response decripted XML structure containing the aditional Mobilpay Order property.
type MResponse struct {
	Order orderM `xml:"order"`
}
type shipping struct {
	SameAsBilling string `xml:"sameasbilling,attr"`
	Fname         string `xml:"first_name"`
	Lname         string `xml:"last_name"`
	FiscalNr      string `xml:"fiscal_number"`
	IdentityNr    string `xml:"identity_number"`
	Country       string `xml:"country"`
	County        string `xml:"county"`
	City          string `xml:"city"`
	ZipCode       string `xml:"zip_code"`
	Address       string `xml:"address"`
	Email         string `xml:"email"`
	MobilePhone   string `xml:"mobile_phone"`
}
type billing struct {
	Type        string `xml:"type,attr"`
	Fname       string `xml:"first_name"`
	Lname       string `xml:"last_name"`
	FiscalNr    string `xml:"fiscal_number"`
	IdentityNr  string `xml:"identity_number"`
	Country     string `xml:"country"`
	County      string `xml:"county"`
	City        string `xml:"city"`
	ZipCode     string `xml:"zip_code"`
	Address     string `xml:"address"`
	Email       string `xml:"email"`
	MobilePhone string `xml:"mobile_phone"`
}
type contactInfo struct {
	Billing  billing  `xml:"billing"`
	Shipping shipping `xml:"shipping"`
}
type invoice struct {
	Currency    string      `xml:"currency,attr"`
	Amount      string      `xml:"amount,attr"`
	Details     string      `xml:"details"`
	ContactInfo contactInfo `xml:"contact_info"`
}
type urlx struct {
	Confirm string `xml:"confirm"`
	Return  string `xml:"return"`
}
type order struct {
	Type      string  `xml:"type,attr"`
	ID        string  `xml:"id,attr"`
	Signature string  `xml:"signature"`
	URL       urlx    `xml:"url"`
	Invoice   invoice `xml:"invoice"`
}
type orderM struct {
	Type      string   `xml:"type,attr"`
	ID        string   `xml:"id,attr"`
	Signature string   `xml:"signature"`
	URL       urlx     `xml:"url"`
	Invoice   invoice  `xml:"invoice"`
	Mobilpay  mobilpay `xml:"mobilpay"`
}

// Request - Mobilpay XML request that gets marshalled by GetOrderEncriptedData
/*	Mandatory fields:
	r.Order.Signature = "xxxx-xxxx-xxxx-xxxx-xxxx"
	r.Order.URL.Return = ""
	r.Order.URL.Confirm = ""
	r.Order.Invoice.ContactInfo.Billing.Type = "person"
	r.Order.Invoice.Currency = "RON"
	r.Order.Invoice.Amount = "3"
	r.Order.Type = "card"
	r.Order.ID = "1"
	r.Order.Invoice.Details = "CreditCard payment."
	r.Order.Invoice.ContactInfo.Billing.Fname = "Foo"
	r.Order.Invoice.ContactInfo.Billing.Lname = "Bar"
*/
type Request struct {
	Order order `xml:"order"`
}

// PrivateKey - parsed
var PrivateKey *rsa.PrivateKey

// Cert - parsed
var Cert *rsa.PublicKey

func (o *CB) insertPaymentStatus(r *http.Request, responseXML MResponse) error {
	id, _ := strconv.Atoi(responseXML.Order.ID)
	b, e := xml.Marshal(responseXML.Order.Mobilpay)
	if e != nil {
		return e
	}
	new := models.PaymentStatus{OrderID: id, ConfirmationXML: string(b), Status: responseXML.Order.Mobilpay.Action}
	if o.DBh != nil {
		tx, e := o.DBh.MasterTx(r.Context(), nil)
		if e != nil {
			log.Println(e.Error())
		}
		if e := new.Insert(r.Context(), tx, boil.Infer()); e != nil {
			tx.Rollback()
			return e
		}
		return tx.Commit()
	}
	return nil
}

// MobilpayConfirm - Confirms and logs the processing of the payment.
func (o *CB) MobilpayConfirm(wr http.ResponseWriter, r *http.Request) {
	log.Println("MobilpayConfirm()::", r.Method, "request")
	err := make([]error, 5)
	responseXML := MResponse{}
	confirmResponse := MerchantsResponse{}
	confirmResponseE := MerchantsResponseE{}
	var conf []byte
	bf := &bytes.Buffer{}
	r.ParseForm()
	log.Println("Form data:", r.Form)
	encKey := r.FormValue("env_key")
	encXML := r.FormValue("data")
	if encKey == "" || encXML == "" {
		log.Println("Invalid parameters received.")
		wr.WriteHeader(http.StatusBadRequest)
		return
	}
	var plain []byte
	plain, _, err[0] = Decrypt(encKey, encXML)
	err[1] = xml.Unmarshal(plain, &responseXML)
	switch responseXML.Order.Mobilpay.Action {
	case "confirmed":
		err[2] = o.insertPaymentStatus(r, responseXML)
		confirmResponse.CRC = "Processed."
		conf, err[3] = xmlMarshal(confirmResponse)
		bf.Write(conf)
		_, err[4] = http.Post(MobilpayEndpoint, "application/xml", bf)
		break
	case "confirmed_pending":
		err[2] = o.insertPaymentStatus(r, responseXML)
		confirmResponse.CRC = "Confirmed pending."
		conf, err[3] = xmlMarshal(confirmResponse)
		bf.Write(conf)
		_, err[4] = http.Post(MobilpayEndpoint, "application/xml", bf)
		break
	case "paid_pending":
		err[2] = o.insertPaymentStatus(r, responseXML)
		confirmResponse.CRC = "Paid pending."
		conf, err[3] = xmlMarshal(confirmResponse)
		bf.Write(conf)
		_, err[4] = http.Post(MobilpayEndpoint, "application/xml", bf)
		break
	case "paid":
		err[2] = o.insertPaymentStatus(r, responseXML)
		confirmResponse.CRC = "Paid."
		conf, err[3] = xmlMarshal(confirmResponse)
		bf.Write(conf)
		_, err[4] = http.Post(MobilpayEndpoint, "application/xml", bf)
		break
	case "canceled":
		err[2] = o.insertPaymentStatus(r, responseXML)
		confirmResponse.CRC = "Canceled."
		conf, err[3] = xmlMarshal(confirmResponse)
		bf.Write(conf)
		_, err[4] = http.Post(MobilpayEndpoint, "application/xml", bf)
		break
	case "credit": // refunded
		err[2] = o.insertPaymentStatus(r, responseXML)
		confirmResponse.CRC = "Credit."
		conf, err[3] = xmlMarshal(confirmResponse)
		bf.Write(conf)
		_, err[4] = http.Post(MobilpayEndpoint, "application/xml", bf)
		break
	default:
		err[2] = o.insertPaymentStatus(r, responseXML)
		confirmResponseE.CRC.ErrorType = 2
		conf, err[3] = xmlMarshal(confirmResponseE)
		bf.Write(conf)
		_, err[4] = http.Post(MobilpayEndpoint, "application/xml", bf)
		log.Printf("Confirm returned with error code %+v", responseXML)
	}
	for k, e := range err {
		if e != nil {
			log.Printf("MobilpayConfirm() error array @ index #%d : %s", k, e.Error())
			wr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	wr.WriteHeader(http.StatusOK)
	return
}

func xmlMarshal(rsp interface{}) ([]byte, error) {
	b, e := xml.Marshal(rsp)
	if e != nil {
		return nil, e
	}
	return b, nil
}

func getPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	key, e := ssh.ParseRawPrivateKey(data)
	return key.(*rsa.PrivateKey), e
}
func getPublicKey(data []byte) (*rsa.PublicKey, error) {
	b, _ := pem.Decode(data)
	cert, e := x509.ParseCertificate(b.Bytes)
	if e != nil {
		log.Println(e.Error())
	}
	pub := cert.PublicKey.(*rsa.PublicKey)
	return pub, e
}

// Encrypt returns encryptedData and encryptedRandomKey base64 encoded
func Encrypt(publicKey *rsa.PublicKey, sourceText []byte) (string, string, error) {
	encryptedText := make([]byte, len(sourceText))
	var ekey []byte
	key := make([]byte, 32)
	rand.Read(key)
	randKey, e := rc4.NewCipher(key)
	if e != nil || publicKey == nil {
		log.Println("Call to encrypt with nil PublicKey")
		return "", "", e
	}
	randKey.XORKeyStream(encryptedText, sourceText)
	ekey, e = rsa.EncryptPKCS1v15(rand.Reader, publicKey, key)
	if e != nil {
		return "", "", e
	}
	encData := base64.StdEncoding.EncodeToString(encryptedText)
	encKey := base64.StdEncoding.EncodeToString(ekey)
	return encData, encKey, nil
}

// Decrypt returns plainTextData and randomKey
func Decrypt(encKey string, encryptedText string) ([]byte, []byte, error) {
	err := make([]error, 4)
	var ekey, etxt, randKey []byte
	cipher := &rc4.Cipher{}
	ekey, err[0] = base64.StdEncoding.DecodeString(encKey)
	etxt, err[1] = base64.StdEncoding.DecodeString(encryptedText)
	sourceText := make([]byte, len(etxt))
	randKey, err[2] = rsa.DecryptPKCS1v15(rand.Reader, PrivateKey, ekey)
	log.Println(string(randKey))
	cipher, err[3] = rc4.NewCipher(randKey)
	cipher.XORKeyStream(sourceText, etxt)
	log.Println(string(sourceText))
	for _, e := range err {
		if e != nil {
			return nil, nil, e
		}
	}
	return sourceText, randKey, nil
}

// EncriptedData - Contains the client data that needs to be POST'ed
type EncriptedData struct {
	EnvKey string `json:"env_key"`
	Data   string `json:"data"`
}

/*
// GetOrderEncriptedData - provides the data that needs to be sent to the client side.
func GetOrderEncriptedData(order *Request, helperObject helper) (EncriptedData, error) {
	data := EncriptedData{}
	var e error
	reqXML := helperObject.xmlMarshal(order)
	data.Data, data.EnvKey, e = Encrypt(cert, reqXML)
	if e != nil {
		return data, e
	}
	return data, nil
}
func post(url, ct string, body io.Reader) *http.Response {
	response, err := http.Post(url, ct, body)
	if err != nil {
		log.Println(err.Error())
	}
	log.Println(response.Header)
	return response
}
func requestData() *Request {
	r := &Request{}
	r.Order.Signature = ""
	r.Order.URL.Return = ""
	r.Order.URL.Confirm = ""
	r.Order.Invoice.ContactInfo.Billing.Type = "person"
	r.Order.Invoice.Currency = "RON"
	r.Order.Invoice.Amount = "3"
	r.Order.Type = "card"
	r.Order.ID = "1"
	r.Order.Invoice.Details = "CreditCard payment."
	r.Order.Invoice.ContactInfo.Billing.Fname = "Foo"
	r.Order.Invoice.ContactInfo.Billing.Lname = "Bar"
	return r
}
func main() {
	cer, _ := ioutil.ReadFile(publicCer)
	cert, _ = getPublicKey(cer)
	b, _ := ioutil.ReadFile(privateKeyFile)
	privateKey, _ = getPrivateKey(b)
	reqXML, e := xml.Marshal(requestData())
	if e != nil {
		log.Println(e.Error())
	}
	etxt, ekey, e := Encrypt(cert, reqXML)
	v := &url.Values{}
	v.Add("env_key", ekey)
	v.Add("data", etxt)
	buf := &bytes.Buffer{}
	buf.WriteString(v.Encode())
	//response := post(mobilpayEndpoint, "application/x-www-form-urlencoded", buf)
	//Decrypt(ekey, etxt)
	//br, e := ioutil.ReadAll(response.Body)
	//log.Println(string(br))
}
*/
