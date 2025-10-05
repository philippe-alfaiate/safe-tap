#!/bin/bash

# Check if the correct number of arguments is provided
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 deviceName profileName"
    exit 1
fi

DEVICE_NAME=$1
PROFILE_NAME=$2

# Generate a timestamp
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Read the public key
PUBLIC_KEY=$(cat "${DEVICE_NAME}_${PROFILE_NAME}-public_key.pem" | awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}' | tr -d '\n')

# Prepare the JSON payload for the curl request
JSON_PAYLOAD=$(jq -n --arg device "$DEVICE_NAME" --arg profile "$PROFILE_NAME" --arg time "$TIMESTAMP" --arg publicKey "$PUBLIC_KEY" \
    '{device: $device, profile: $profile, date: $time, public_key: $publicKey}')
echo -n "$JSON_PAYLOAD" > payload.json
# Sign the payload
$(openssl dgst -sign  ${DEVICE_NAME}_${PROFILE_NAME}-private_key.pem -keyform PEM -sha256 -out sign.bin -binary payload.json)

SIGNATURE=$(base64 sign.bin | tr -d '\n')

# Add the signature to the header
HEADER="X-sign: $SIGNATURE"
echo "Signature: $SIGNATURE"

# Make the curl request to verify the message
curl -k -X POST https://localhost:9809/api/v0/verify \
    -H "Content-Type: application/json" \
    -H "$HEADER" \
    -d "$JSON_PAYLOAD" \
    -o response.json