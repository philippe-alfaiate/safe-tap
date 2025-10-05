#!/bin/bash

# Check if the correct number of arguments is provided
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 deviceName profileName"
    exit 1
fi

DEVICE_NAME=$1
PROFILE_NAME=$2

# Generate ECDSA private key
openssl ecparam -genkey -name prime256v1 -noout -out "${DEVICE_NAME}_${PROFILE_NAME}-private_key.pem"

# Generate ECDSA public key
openssl ec -in "${DEVICE_NAME}_${PROFILE_NAME}-private_key.pem" -pubout -out "${DEVICE_NAME}_${PROFILE_NAME}-public_key.pem"

# Read the public key
PUBLIC_KEY=$(cat ${DEVICE_NAME}_${PROFILE_NAME}-public_key.pem | awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}' | tr -d '\n')

# Prepare the JSON payload for the curl request
JSON_PAYLOAD=$(jq -n --arg name "$DEVICE_NAME" --arg public_key "$PUBLIC_KEY" --argjson profiles "[\"$PROFILE_NAME\"]" \
    '{name: $name, public_key: $public_key, profiles: $profiles}')

# Make the curl request to register the device
curl -k -X POST https://localhost:9809/api/v0/device/add \
    -H "Content-Type: application/json" \
    -d "$JSON_PAYLOAD" \
    -o response.json
