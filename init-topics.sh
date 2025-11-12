#!/bin/bash

echo "Waiting for Kafka to be ready."
sleep 10

# Cart aggregate events topic
kafka-topics --create --if-not-exists \
  --bootstrap-server kafka:9092 \
  --topic ec.cart-events \
  --partitions 3 \
  --replication-factor 1

# Misc events topic (fallback)
kafka-topics --create --if-not-exists \
  --bootstrap-server kafka:9092 \
  --topic ec.misc-events \
  --partitions 3 \
  --replication-factor 1

echo "Topics created successfully:"
kafka-topics --list --bootstrap-server kafka:9092
