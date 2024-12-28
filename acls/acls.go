package acls

import (
	"fmt"
	"strings"

	"github.com/rca0/kafka-control/config"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
)

type ACLConfig struct {
	Roles map[string][]ACL `yaml:"roles"`
	Users map[string]User  `yaml:"users"`
}
type ACL struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"`
	Pattern    string `yaml:"pattern"`
	Host       string `yaml:"host"`
	Operation  string `yaml:"operation"`
	Permission string `yaml:"permission"`
}

type User struct {
	Principal string   `yaml:"principal"`
	Roles     []string `yaml:"roles"`
}

func ManageACLs(cfg *config.Config, aclConfig ACLConfig, plan bool) error {
	tlsConfig, saslMechanism, err := cfg.TLSConfig()
	if err != nil {
		return err
	}

	transport := &kafka.Transport{TLS: tlsConfig}

	if saslMechanism != nil {
		transport.SASL = sasl.Mechanism(*saslMechanism)
	}

	client := &kafka.Client{
		Addr:      kafka.TCP(cfg.Brokers),
		Transport: transport,
	}

	for _, user := range aclConfig.Users {
		for _, role := range user.Roles {
			acls := aclConfig.Roles[role]
			if plan {
				for _, v := range acls {
					fmt.Printf("-----------------------\n")
					fmt.Printf("[PLAN] Create ACL\n")
					fmt.Printf("  -  Name: %s\n", v.Name)
					fmt.Printf("  -  Type: %s\n", v.Type)
					fmt.Printf("  -  Principal: %d\n", user.Principal)
					fmt.Printf("  -  Pattern: %s\n", v.Pattern)
					fmt.Printf("  -  Host: %s\n", v.Host)
					fmt.Printf("  -  Operation: %s\n", v.Operation)
					fmt.Printf("  -  Permission: %s\n", v.Permission)
				}
			} else {
				if err := createACL(client, acls, user.Principal); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func createACL(client *kafka.Client, acls []ACL, principal string) error {
	var reqACLs []kafka.ACLEntry

	for _, acl := range acls {
		var (
			resourceType kafka.ResourceType
			patternType  kafka.PatternType
			operation    kafka.ACLOperationType
			permission   kafka.ACLPermissionType
		)

		switch strings.ToUpper(acl.Type) {
		case "TOPIC":
			resourceType = kafka.ResourceTypeTopic
		case "GROUP":
			resourceType = kafka.ResourceTypeGroup
		case "CLUSTER":
			resourceType = kafka.ResourceTypeCluster
		default:
			return fmt.Errorf("unknown resource type; %s", acl.Type)
		}

		switch strings.ToUpper(acl.Pattern) {
		case "LITERAL":
			patternType = kafka.PatternTypeLiteral
		case "PREFIXED":
			patternType = kafka.PatternTypePrefixed
		default:
			return fmt.Errorf("unknown pattern type; %s", acl.Pattern)
		}

		switch strings.ToUpper(acl.Operation) {
		case "ALL":
			operation = kafka.ACLOperationTypeAll
		case "READ":
			operation = kafka.ACLOperationTypeRead
		case "WRITE":
			operation = kafka.ACLOperationTypeWrite
		case "CREATE":
			operation = kafka.ACLOperationTypeCreate
		case "DELETE":
			operation = kafka.ACLOperationTypeDelete
		case "ALTER":
			operation = kafka.ACLOperationTypeAlter
		case "DESCRIBE":
			operation = kafka.ACLOperationTypeDescribe
		case "DESCRIBE_CONFIGS":
			operation = kafka.ACLOperationTypeDescribeConfigs
		case "CLUSTER_ACTION":
			operation = kafka.ACLOperationTypeClusterAction
		case "ALTER_CONFIGS":
			operation = kafka.ACLOperationTypeAlterConfigs
		case "IDEMPOTENT_WRITE":
			operation = kafka.ACLOperationTypeIdempotentWrite
		default:
			return fmt.Errorf("unknown operation type; %s", acl.Operation)
		}

		switch strings.ToUpper(acl.Permission) {
		case "ALLOW":
			permission = kafka.ACLPermissionTypeAllow
		case "DENY":
			permission = kafka.ACLPermissionTypeDeny
		case "ANY":
			permission = kafka.ACLPermissionTypeAny
		default:
			fmt.Errorf("unknown permission type; %s", acl.Permission)
		}

		reqACLs = append(reqACLs, kafka.ACLEntry{
			ResourceType: resourceType,
			ResourceName: acl.Name,
			ResourcePatternType: patternType,
			Principal: principal,
			Host: acl.Host,
			Operation: operation,
			PermissionType: permission,
		}))
	}

	req := &kafka.CreateACLsrequest{
		ACLs: reqACLs,
	}

	res, err := client.CreateACLs(context.Background(),req)
	if err != nil {
		return err 
	}

	for _, aclErr := range res.Errors {
		if aclErr != nil {
			return aclErr
		}
	}

	for _, acl := range acls {
		fmt.Printf("-----------------------\n")
		fmt.Printf("[APPLY] Created ACL\n")
		fmt.Printf("  -  Name: %s\n", acl.Name)
		fmt.Printf("  -  Type: %s\n", acl.Type)
		fmt.Printf("  -  Principal: %d\n", principal)
		fmt.Printf("  -  Pattern: %s\n", acl.Pattern)
		fmt.Printf("  -  Host: %s\n", acl.Host)
		fmt.Printf("  -  Operation: %s\n", acl.Operation)
		fmt.Printf("  -  Permission: %s\n", acl.Permission)
	}
	return nil
}
