# kaf-cfg

CLI-tool to manage Kafka ACLs and Topics.

## Env vars

```bash
export KAFKA_BROKERS="rca0.domain.com:9093"
export KAFKA_SSL_ENABLED=true
export KAFKA_SSL_CA_CONTENT=$(cat /userhome/rca0/ca.crt)
export KAFKA_SSL_CERT_CONTENT=$(cat /userhome/rca0/tls.crt)
export KAFKA_SSL_KEY_CONTENT=$(cat /userhome/rca0/tls.key)
export KAFKA_SASL_ENABLED=true
export KAFKA_SASL_USERNAME=admin
export KAFKA_SASL_PASSWORD=super-secret
export KAFKA_SASL_MECHANISM=PLAIN
```