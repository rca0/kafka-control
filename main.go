package main

import (
	"flag"
	"log"
	"os"

	"github.com/rca0/kaf-cfg/acls"
	"github.com/rca0/kaf-cfg/config"
	"github.com/rca0/kaf-cfg/topics"
	"gopkg.in/yaml.v2"
)

func main() {
	var (
		cmdFlag  = flag.String("cmd", "", "command to execute; topics or acls")
		planPlag = flag.Bool("plan", false, "plan mode to compare the configuration without applyting changes")
		fileFlag = flag.String("f", "", "path to the yaml configuration file")
	)
	flag.Parse()

	if *cmdFlag == "" || *fileFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(*fileFlag)
	if err != nil {
		log.Fatalf("failed to load configuration; %s", err)
	}

	data, err := os.ReadFile(*fileFlag)
	if err != nil {
		log.Fatalf("error reading file %v\n", err)
	}

	switch *cmdFlag {
	case "topics":
		var topicConfig topics.TopicConfig
		if err := yaml.Unmarshal(data, &topicConfig); err != nil {
			log.Fatalf("error unmarshalling topic config; %v\n", err)
		}

		if err := topics.ManageTopics(cfg, topicConfig, *planPlag); err != nil {
			log.Fatalf("failed to manage topics; %v", err)
		}
	case "acls":
		var aclConfig acls.ACLConfig
		if err := yaml.Unmarshal(data, &aclConfig); err != nil {
			log.Fatalf("error unmarshalling ACL config; %v\n", err)
		}

		if err := acls.ManageACLs(cfg, aclConfig, *planPlag); err != nil {
			log.Fatalf("failed to manage ACLS: %v", err)
		}
	default:
		log.Printf("invalid command. Supported commands; topics, acls")
		os.Exit(1)
	}
}
