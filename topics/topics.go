package topics

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/rca0/kafka-control/config"
	"github.com/segmentio/kafka-go"
)

type TopicConfig struct {
	Topics map[string]TopicDetail `yaml:"topics"`
}

type TopicDetail struct {
	Partitions int               `yaml:"partitions"`
	Replicas   int               `yaml:"replicas"`
	Config     map[string]string `yaml:"config"`
}

func ManageTopics(cfg *config.Config, topicConfig TopicConfig, plan bool) error {
	dialer, err := cfg.KafkaDialer()
	if err != nil {
		return err
	}

	conn, err := dialer.DialContext(context.Background(), "tcp", cfg.Brokers)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	var controllerConn *kafka.Conn
	controllerConn, err = dialer.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	for topicName, details := range topicConfig.Topics {
		if plan {
			fmt.Printf("-----------------------\n")
			fmt.Printf("[PLAN] Create Topic\n")
			fmt.Printf("  -  Topic: %s\n", topicName)
			fmt.Printf("  -  Partitions: %d\n", details.Partitions)
			fmt.Printf("  -  Replicas: %d\n", details.Replicas)
			if len(details.Config) > 0 {
				fmt.Printf("  -  Config: \n")
				for k, v := range details.Config {
					fmt.Printf("  -  %s: %s", k, v)
				}
			}
		} else {
			if err := createTopics(controllerConn, topicName, details); err != nil {
				return err
			}
		}
	}
	return nil
}

func createTopics(conn *kafka.Conn, topicName string, details TopicDetail) error {
	topicConfig := kafka.TopicConfig{
		Topic:             topicName,
		NumPartitions:     details.Partitions,
		ReplicationFactor: details.Replicas,
		ConfigEntries:     make([]kafka.ConfigEntry, 0, len(details.Config)),
	}

	for k, v := range details.Config {
		topicConfig.ConfigEntries = append(topicConfig.ConfigEntries, kafka.ConfigEntry{
			ConfigName:  k,
			ConfigValue: v,
		})
	}

	err := conn.CreateTopics(topicConfig)
	if err != nil {
		log.Printf("failed to create topic; %v\n", err)
		return err
	}

	fmt.Printf("-----------------------\n")
	fmt.Printf("[APPY] Created Topic\n")
	fmt.Printf("  -  Topic: %s\n", topicName)
	fmt.Printf("  -  Partitions: %d\n", details.Partitions)
	fmt.Printf("  -  Replicas: %d\n", details.Replicas)
	if len(topicConfig.ConfigEntries) > 0 {
		fmt.Printf("  -  Config: \n")
		for _, v := range topicConfig.ConfigEntries {
			fmt.Printf("  -  %s", v)
		}
	}
	return nil
}
