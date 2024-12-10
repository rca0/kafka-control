package config

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type Config struct {
	Brokers string
	SSL     SSLConfig
	SASL    SASLConfig
}

type SSLConfig struct {
	Enabled     bool
	KeyPassword string
	CAContent   string
	CertContent string
	KeyContent  string
}

type SASLConfig struct {
	Enabled   bool
	Username  string
	Password  string
	Mechanism string // PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
}

func LoadConfig(file string) (*Config, error) {
	log.Println("load configs")

	var cfg Config
	envVars := map[string]*string{
		"KAFKA_BROKERS":          &cfg.Brokers,
		"KAFKA_SSL_KEY_PASSWORD": &cfg.SSL.KeyPassword,
		"KAFKA_SASL_USERNAME":    &cfg.SASL.Username,
		"KAFKA_SASL_PASSWORD":    &cfg.SASL.Password,
		"KAFKA_SASL_MECHANISM":   &cfg.SASL.Mechanism,
		"KAFKA_SSL_KEY_CONTENT":  &cfg.SSL.KeyContent,
		"KAFKA_SSL_CA_CONTENT":   &cfg.SSL.CAContent,
		"KAFKA_SSL_CERT_CONTENT": &cfg.SSL.CertContent,
	}

	for env, field := range envVars {
		if value := os.Getenv(env); value != "" {
			*field = value
		}
	}

	boolEnvVars := map[string]*bool{
		"KAFKA_SSL_ENABLED":  &cfg.SSL.Enabled,
		"KAFKA_SASL_ENABLED": &cfg.SASL.Enabled,
	}

	for env, field := range boolEnvVars {
		if value := os.Getenv(env); value != "" {
			*field = value == "true"
		}
	}

	return &cfg, nil
}

func (cfg *Config) KafkaDialer() (*kafka.Dialer, error) {
	dialer := &kafka.Dialer{}

	if cfg.SASL.Enabled {
		mechanism := plain.Mechanism{
			Username: cfg.SASL.Username,
			Password: cfg.SASL.Password,
		}

		dialer.SASLMechanism = mechanism
	}

	if cfg.SSL.Enabled {
		certContent := []byte(cfg.SSL.CertContent)
		keyContent := []byte(cfg.SSL.KeyContent)
		caContent := []byte(cfg.SSL.CAContent)

		cert, err := tls.X509KeyPair(certContent, keyContent)
		if err != nil {
			log.Printf("failed to load client certificate/key pair file; %v", err)
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caContent)

		dialer.TLS = &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
		}
	}
	return dialer, nil
}

func (cfg *Config) TLSConfig() (*tls.Config, *plain.Mechanism, error) {
	certContent := []byte(cfg.SSL.CertContent)
	keyContent := []byte(cfg.SSL.KeyContent)
	caContent := []byte(cfg.SSL.CAContent)

	cert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		log.Printf("failed to load client certificate/key pair file; %v", err)
		return nil, nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caContent)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
	}

	var saslMechanism *plain.Mechanism

	if cfg.SASL.Enabled {
		mechanism := plain.Mechanism{
			Username: cfg.SASL.Username,
			Password: cfg.SASL.Password,
		}
		saslMechanism = &mechanism
	}
	return tlsConfig, saslMechanism, nil
}
