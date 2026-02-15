package com.inventium.inventory_transaction_service.controller;

import com.inventium.inventory_transaction_service.dto.TransactionRecordRequest;
import com.inventium.inventory_transaction_service.dto.TransactionRecordResponse;
import com.inventium.inventory_transaction_service.service.TransactionRecordService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.PathVariable;


@RestController
@RequestMapping("/api/v1/transactions")
@RequiredArgsConstructor
public class TransactionRecordController {
    
    private final TransactionRecordService transactionRecordService;
    
    @PostMapping
    public ResponseEntity<TransactionRecordResponse> createTransaction(
            @Valid @RequestBody TransactionRecordRequest request) {
        TransactionRecordResponse response = transactionRecordService.createTransactionRecord(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }
    
    @GetMapping("/{inventoryId}/{id}")
    public ResponseEntity<TransactionRecordResponse> getTransaction(
            @PathVariable String inventoryId, 
            @PathVariable String id) {
        TransactionRecordResponse response = transactionRecordService.getTransactionRecord(inventoryId, id);
        return ResponseEntity.ok(response);
    }

    @PutMapping("/{inventoryId}/{id}")
    public ResponseEntity<TransactionRecordResponse> updateTransaction(
        @Valid @RequestBody TransactionRecordRequest request,
        @PathVariable String inventoryId,
        @PathVariable String id) {
        TransactionRecordResponse response = transactionRecordService.updateTransactionRecord(request, inventoryId, id);
        return ResponseEntity.ok(response);
        }
}
