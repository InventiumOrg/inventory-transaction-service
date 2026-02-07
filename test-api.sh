#!/bin/bash

BASE_URL="http://localhost:14330"

echo "=== Testing Inventory Transaction Service API ==="

echo -e "\n1. Health Check:"
curl -s -X GET $BASE_URL/health | jq

echo -e "\n2. Create Import Transaction:"
RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "type": "IMPORT",
    "source": "Warehouse 11",
    "destination": "POS 02",
    "inventoryId": "49",
    "inventoryMeasure": "bag",
    "inventoryCategory": "Coffee",
    "inventoryUnit": "kg",
    "quantity": 10
  }')

echo $RESPONSE | jq
TRANSACTION_ID=$(echo $RESPONSE | jq -r '.id')
INVENTORY_ID=$(echo $RESPONSE | jq -r '.inventoryId')

echo -e "\n3. Get Transaction by inventoryId and ID:"
curl -s -X GET $BASE_URL/api/v1/transactions/$INVENTORY_ID/$TRANSACTION_ID | jq

echo -e "\n4. Update Import Transaction Status to Completed:"
RESPONSE=$(curl -s -X PUT $BASE_URL/api/v1/transactions/$INVENTORY_ID/$TRANSACTION_ID \
  -H "Content-Type: application/json" \
  -d '{
    "type": "IMPORT",
    "source": "Warehouse 11",
    "destination": "POS 02",
    "inventoryId": "49",
    "inventoryMeasure": "bag",
    "inventoryCategory": "Coffee",
    "inventoryUnit": "kg",
    "quantity": 10,
    "status": "Completed"
  }')
echo $RESPONSE | jq

echo -e "\n5. Test Validation Error:"
curl -s -X POST $BASE_URL/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "type": "",
    "quantity": -5
  }' | jq

echo -e "\n6. Test Not Found Error:"
curl -s -X GET $BASE_URL/api/v1/transactions/non-existent-inventory/non-existent-id | jq

echo -e "\n=== API Testing Complete ==="