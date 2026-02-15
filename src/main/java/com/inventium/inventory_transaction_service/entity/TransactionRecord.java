package com.inventium.inventory_transaction_service.entity;

import software.amazon.awssdk.enhanced.dynamodb.mapper.annotations.DynamoDbBean;
import software.amazon.awssdk.enhanced.dynamodb.mapper.annotations.DynamoDbPartitionKey;
import software.amazon.awssdk.enhanced.dynamodb.mapper.annotations.DynamoDbSortKey;

import java.time.Instant;

import lombok.*;

@Getter @Setter @NoArgsConstructor 

@DynamoDbBean
public class TransactionRecord {
    private String id;
    private String type;
    private String source;
    private String destination;
    private Instant createdAt = Instant.now();
    private String inventoryId;
    private String inventoryMeasure;
    private String inventoryCategory;
    private String inventoryUnit;
    private Integer quantity;
    private String status;

    @DynamoDbSortKey
    public String getId(){
        return id;
    }
    @DynamoDbPartitionKey
    public String getInventoryId() {
        return inventoryId;
    }
}
