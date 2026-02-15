package com.inventium.inventory_transaction_service.dto;

import lombok.Data;
import java.time.Instant;

@Data
public class TransactionRecordResponse {
    private String id;
    private String type;
    private String source;
    private String destination;
    private Instant createdAt;
    private String inventoryId;
    private String inventoryMeasure;
    private String inventoryCategory;
    private String inventoryUnit;
    private Integer quantity;
    private String status;
}