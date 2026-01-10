package com.inventium.inventory_transaction_service.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import lombok.Data;

@Data
public class TransactionRecordRequest {
    @NotBlank(message = "Type is required")
    private String type; // "IMPORT" or "EXPORT"
    
    @NotBlank(message = "Source is required")
    private String source;
    
    @NotBlank(message = "Destination is required")
    private String destination;
    
    @NotBlank(message = "Inventory ID is required")
    private String inventoryId;
    
    @NotBlank(message = "Inventory measure is required")
    private String inventoryMeasure;
    
    @NotBlank(message = "Inventory category is required")
    private String inventoryCategory;
    
    @NotBlank(message = "Inventory unit is required")
    private String inventoryUnit;
    
    @NotNull(message = "Quantity is required")
    @Positive(message = "Quantity must be positive")
    private Integer quantity;

    private String status;
}