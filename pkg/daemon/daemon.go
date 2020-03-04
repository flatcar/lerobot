package daemon

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-acme/lego/v3/certcrypto"
	legocert "github.com/go-acme/lego/v3/certificate"
	"github.com/go-acme/lego/v3/lego"
	legolog "github.com/go-acme/lego/v3/log"
	"github.com/go-acme/lego/v3/providers/dns/route53"
	"github.com/go-acme/lego/v3/registration"
	yaml "gopkg.in/yaml.v2"

	"github.com/kinvolk/lerobot/pkg/util"
)

func init() {
	legolog.Logger = log.New(os.Stderr, "[lego] ", log.LstdFlags)
}

type Daemon struct {
	mutex sync.RWMutex

	doneChan chan struct{}

	options *Options

	accounts     []Account
	certificates []Certificate
}

type Options struct {
	Interval           time.Duration
	LEConfigPath       string
	LEAPI              string
	AccountDir         string
	CertificateDir     string
	AuthorizedKeysPath string
}

type Account struct {
	Email        string `yaml:"email"`
	SSHPublicKey string `yaml:"ssh_public_key"`

	dir string
}

type Certificate struct {
	AccountEmail string   `yaml:"account"`
	CommonName   string   `yaml:"common_name"`
	SAN          []string `yaml:"subject_alternative_names"`

	dir string
}

type Configuration struct {
	Accounts     []Account     `yaml:"accounts"`
	Certificates []Certificate `yaml:"certificates"`
}

// legoAccount implements acme.User as required by lego.
// https://go-acme.github.io/lego/usage/library/
type legoAccount struct {
	Email string `json:"email"`

	Registration *registration.Resource `json:"registration"`

	privateKey crypto.PrivateKey
}

func (a *legoAccount) GetEmail() string {
	return a.Email
}

func (a *legoAccount) GetRegistration() *registration.Resource {
	return a.Registration
}

func (a *legoAccount) GetPrivateKey() crypto.PrivateKey {
	return a.privateKey
}

func loadLegoAccount(accountDir, email string) (*legoAccount, error) {
	accountDir = path.Join(accountDir, email)
	if err := os.MkdirAll(accountDir, 0700); err != nil {
		return nil, err
	}

	accountFilePath := path.Join(accountDir, "account.json")

	privateKeyPath := path.Join(accountDir, "private.pem")

	var privateKey crypto.PrivateKey

	_, err := os.Stat(privateKeyPath)
	if os.IsNotExist(err) {
		privateKey, err = generatePrivateKey(privateKeyPath)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		privateKey, err = loadPrivateKey(privateKeyPath)
		if err != nil {
			return nil, err
		}
	}

	_, err = os.Stat(accountFilePath)
	if os.IsNotExist(err) {
		return &legoAccount{
			Email:      email,
			privateKey: privateKey,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Loading account file %q", accountFilePath)

	accountBytes, err := ioutil.ReadFile(accountFilePath)
	if err != nil {
		return nil, err
	}

	var account legoAccount
	if err := json.Unmarshal(accountBytes, &account); err != nil {
		return nil, err
	}

	account.privateKey = privateKey

	if account.Registration == nil {
		return nil, fmt.Errorf("cannot load account, registration is nil")
	}

	return &account, nil
}

func loadLegoClient(config *lego.Config) (*lego.Client, error) {
	legoClient, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	provider, err := route53.NewDNSProvider()
	if err != nil {
		return nil, err
	}

	if err := legoClient.Challenge.SetDNS01Provider(provider); err != nil {
		return nil, err
	}

	return legoClient, nil
}

func writeCert(certificateDir string, certRes *legocert.Resource) {
	if err := os.MkdirAll(certificateDir, 0700); err != nil {
		log.Printf("[ERROR] Unable to create certificate dir %q: %v", certificateDir, err)
	}

	certificateFile := path.Join(certificateDir, certRes.Domain+".crt")
	privateKeyFile := path.Join(certificateDir, certRes.Domain+".key")
	metadataFile := path.Join(certificateDir, certRes.Domain+".json")
	issueCertificateFile := path.Join(certificateDir, certRes.Domain+".issuer.crt")

	if err := util.WriteFileAtomic(certificateFile, certRes.Certificate, 0600); err != nil {
		log.Printf("[ERROR] Unable to save Certificate for domain %q\n\t%v", certRes.Domain, err)
		return
	}

	if certRes.IssuerCertificate != nil {
		if err := util.WriteFileAtomic(issueCertificateFile, certRes.IssuerCertificate, 0600); err != nil {
			log.Printf("[ERROR] Unable to save IssuerCertificate for domain %q\n\t%v", certRes.Domain, err)
			return
		}
	}

	if err := util.WriteFileAtomic(privateKeyFile, certRes.PrivateKey, 0600); err != nil {
		log.Printf("[ERROR] Unable to save PrivateKey for domain %q\n\t%v", certRes.Domain, err)
		return
	}

	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		log.Printf("[ERROR] Unable to marshal CertificateResource for domain %q\n\t%v", certRes.Domain, err)
		return
	}

	if err := util.WriteFileAtomic(metadataFile, jsonBytes, 0600); err != nil {
		log.Printf("[ERROR] Unable to save CertResource for domain %q\n\t%v", certRes.Domain, err)
	}
}

func (d *Daemon) loadConfig() error {
	configBytes, err := ioutil.ReadFile(d.options.LEConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %q: %v", d.options.LEConfigPath, err)
	}

	var configuration Configuration
	if err := yaml.Unmarshal(configBytes, &configuration); err != nil {
		return fmt.Errorf("failed to unmarshal config: %v", err)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.accounts = configuration.Accounts
	d.certificates = configuration.Certificates

	return nil
}

const (
	sshAuthKeysTmpl = "command=\"rsync --server --sender -vlogDtpre.iLsfxC . certificates/%s/\" %s\n"
)

func (d *Daemon) updateAuthorizedKeys() {
	if d.options.AuthorizedKeysPath == "" {
		log.Printf("[INFO] No authorized-keys-file configuration, not updating")
		return
	}
	targetDir := filepath.Dir(d.options.AuthorizedKeysPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Printf("[ERROR] Unable to create dir %q: %v", targetDir, err)
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	var authKeys bytes.Buffer
	for _, account := range d.accounts {
		_, err := authKeys.WriteString(fmt.Sprintf(sshAuthKeysTmpl, account.Email, account.SSHPublicKey))
		if err != nil {
			log.Printf("[ERROR] Failed to write into buffer: %v", err)
			continue
		}
	}

	if err := util.WriteFileAtomic(d.options.AuthorizedKeysPath, authKeys.Bytes(), 0644); err != nil {
		log.Printf("[ERROR] Failed to write new authorized_keys file: %v", err)
	} else {
		log.Printf("[INFO] Updated %q file", d.options.AuthorizedKeysPath)
	}
}

func New(options *Options) (*Daemon, error) {
	daemon := &Daemon{
		doneChan: make(chan struct{}),
		options:  options,
	}
	if err := daemon.loadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config %q: %v", daemon.options.LEConfigPath, err)
	}

	return daemon, nil
}

func (d *Daemon) Run() {
	log.Printf("[INFO] Using Let's Encrypt API endpoint: %s", d.options.LEAPI)
	d.updateAuthorizedKeys()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	go func() {
		for {
			select {
			case <-d.doneChan:
				return
			case <-sigChan:
			}
			if err := d.loadConfig(); err != nil {
				log.Printf("[ERROR] Failed to reload config: %v")
			} else {
				log.Printf("[INFO] Configuration from %q loaded", d.options.LEConfigPath)
			}
			d.updateAuthorizedKeys()
		}
	}()

	for {
		timer := time.NewTimer(d.options.Interval)

		d.requestCertificates()
		d.renewCertificates()

		select {
		case <-d.doneChan:
			return
		default:
		}
		select {
		case <-timer.C:
			continue
		case <-d.doneChan:
			return
		}
	}
}

func saveAccount(accountDir string, account *legoAccount) error {
	jsonBytes, err := json.MarshalIndent(account, "", "\t")
	if err != nil {
		return err
	}
	return util.WriteFileAtomic(path.Join(accountDir, account.Email, "account.json"), jsonBytes, 0600)
}

func registerAccountAcceptTOS(accountDir string, account *legoAccount, client *lego.Client) error {
	if account.Registration == nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return err
		}
		account.Registration = reg
		return saveAccount(accountDir, account)
	}
	return nil
}

func (d *Daemon) requestCertificates() {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	for _, certificate := range d.certificates {
		d.requestCertificate(certificate, false)
	}
}

func (d *Daemon) requestCertificate(certificate Certificate, force bool) {
	log.Printf("[INFO] Requesting certificate for %q", certificate.CommonName)

	certificatePath := path.Join(d.options.CertificateDir, certificate.AccountEmail, certificate.CommonName+".crt")
	_, err := os.Stat(certificatePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("[ERROR] Failed to stat %q: %v", certificatePath, err)
			return
		}
	} else if !force {
		// Nothing to do. Certificate exists already and will
		// be renewed later if necessary
		log.Printf("[INFO] Certificate for %q exists already, nothing to do", certificate.CommonName)
		return
	}

	legoAccount, err := loadLegoAccount(d.options.AccountDir, certificate.AccountEmail)
	if err != nil {
		log.Printf("[ERROR] Failed to load account for %q: %v", certificate.AccountEmail, err)
		return
	}

	legoConfig := lego.NewConfig(legoAccount)

	legoConfig.CADirURL = d.options.LEAPI
	legoConfig.Certificate.KeyType = certcrypto.RSA4096

	legoClient, err := loadLegoClient(legoConfig)
	if err != nil {
		log.Printf("[ERROR] Failed to load lego client: %v", err)
		return
	}

	if err := registerAccountAcceptTOS(d.options.AccountDir, legoAccount, legoClient); err != nil {
		log.Printf("[ERROR] Failed to register account %q and accept TOS: %v", legoAccount.Email, err)
		return
	}

	nameList := []string{
		certificate.CommonName,
	}
	nameList = append(nameList, certificate.SAN...)

	request := legocert.ObtainRequest{
		Domains: nameList,
	}

	cert, err := legoClient.Certificate.Obtain(request)
	if err != nil {
		log.Printf("[ERROR] Failed to obtain certificate for %q: %v", certificate.CommonName, err)
		return
	}

	writeCert(path.Join(d.options.CertificateDir, legoAccount.Email), cert)
}

func (d *Daemon) renewCertificates() {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	for _, certificate := range d.certificates {
		d.renewCertificate(certificate)
	}
}

func (d *Daemon) renewCertificate(certificate Certificate) {
	certificatePath := path.Join(d.options.CertificateDir, certificate.AccountEmail, certificate.CommonName+".crt")

	_, err := os.Stat(certificatePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[WARN] Certificate for %q doesn't exist, cannot renew", certificate.CommonName)
		} else {
			log.Printf("[ERROR] Failed to stat %q: %v", certificatePath, err)
		}
		return
	}

	pemBytes, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		log.Printf("[ERROR] Failed to read certificate file %q: %v", certificatePath, err)
		return
	}

	pemBlock, _ := pem.Decode([]byte(pemBytes))

	currentCert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		log.Printf("[ERROR] Failed to parse certificate %q: %v", certificatePath, err)
		return
	}

	currentCommonName := strings.TrimPrefix(currentCert.Subject.String(), "CN=")

	currentNames := []string{
		currentCommonName,
	}
	currentNames = append(currentNames, currentCert.DNSNames...)

	requestedNames := []string{
		certificate.CommonName,
	}
	// The Common Name also is included in SANs
	requestedNames = append(requestedNames, certificate.CommonName)
	requestedNames = append(requestedNames, certificate.SAN...)

	if !equalNameLists(currentNames, requestedNames) {
		// We already have a certificate with that common name
		// but the list of Subject Alternative Names changed,
		// therefore we request a new certificate
		log.Printf("[INFO] Requesting new certificate for %q since SAN list changed", currentCommonName)
		d.requestCertificate(certificate, true)
		return
	}

	validDays := int(currentCert.NotAfter.Sub(time.Now()).Hours() / 24)
	if validDays > 30 {
		log.Printf("[INFO] Certificate %q is valid for %d more days, not renewing", currentCommonName, validDays)
		return
	}

	log.Printf("[INFO] Renewing certificate %q since only valid for %d more days", currentCommonName, validDays)
	d.requestCertificate(certificate, true)
}

func (d *Daemon) Shutdown(_ context.Context) error {
	// TODO(schu): Once we use an ACME library that supports
	// context, we can actively cancel current operations from
	// here
	select {
	case <-d.doneChan:
		// already closed
	default:
		close(d.doneChan)
	}
	return nil
}

func equalNameLists(a, b []string) bool {
	// log.Printf("[DEBUG] Comparing name lists:\n%v\n%v", a, b)
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
