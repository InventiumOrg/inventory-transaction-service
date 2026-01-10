package com.inventium.inventory_transaction_service.service;

import com.inventium.inventory_transaction_service.dto.TransactionRecordRequest;
import com.inventium.inventory_transaction_service.dto.TransactionRecordResponse;
import com.inventium.inventory_transaction_service.entity.TransactionRecord;
import com.inventium.inventory_transaction_service.repository.TransactionRecordRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import com.inventium.inventory_transaction_service.exception.TransactionNotFoundException;

import com.inventium.inventory_transaction_service.processor.InventoryProducer;

import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class TransactionRecordService {
    
    private final TransactionRecordRepository transactionRecordRepository;
    private final InventoryProducer inventoryProducer;
    // private final KafkaProducerService kafkaProducerService;
    
    public TransactionRecordResponse createTransactionRecord(TransactionRecordRequest request) {
        log.info("Creating transaction record for type: {}", request.getType());
        
        TransactionRecord transactionRecord = new TransactionRecord();
        transactionRecord.setId(UUID.randomUUID().toString());
        transactionRecord.setType(request.getType());
        transactionRecord.setSource(request.getSource());
        transactionRecord.setDestination(request.getDestination());
        transactionRecord.setInventoryId(request.getInventoryId());
        transactionRecord.setInventoryMeasure(request.getInventoryMeasure());
        transactionRecord.setInventoryCategory(request.getInventoryCategory());
        transactionRecord.setInventoryUnit(request.getInventoryUnit());
        transactionRecord.setQuantity(request.getQuantity());
        transactionRecord.setStatus("In Progress");
        
        TransactionRecord savedRecord = transactionRecordRepository.save(transactionRecord);
        
        log.info("Transaction record created with ID: {}", savedRecord.getId());
        
        return mapToResponse(savedRecord);
    }
    
    public TransactionRecordResponse getTransactionRecord(String inventoryId, String id) {
        log.info("Retrieving transaction record with inventoryId: {} and ID: {}", inventoryId, id);
        
        TransactionRecord transactionRecord = transactionRecordRepository.findByIdOptional(inventoryId, id)
            .orElseThrow(() -> new TransactionNotFoundException("Transaction record not found with inventoryId: " + inventoryId + " and ID: " + id));
        
        return mapToResponse(transactionRecord);
    }
    
    private TransactionRecordResponse mapToResponse(TransactionRecord transactionRecord) {
        TransactionRecordResponse response = new TransactionRecordResponse();
        response.setId(transactionRecord.getId());
        response.setType(transactionRecord.getType());
        response.setSource(transactionRecord.getSource());
        response.setDestination(transactionRecord.getDestination());
        response.setCreatedAt(transactionRecord.getCreatedAt());
        response.setInventoryId(transactionRecord.getInventoryId());
        response.setInventoryMeasure(transactionRecord.getInventoryMeasure());
        response.setInventoryCategory(transactionRecord.getInventoryCategory());
        response.setInventoryUnit(transactionRecord.getInventoryUnit());
        response.setQuantity(transactionRecord.getQuantity());
        response.setStatus(transactionRecord.getStatus());
        return response;
    }

    public TransactionRecordResponse updateTransactionRecord(TransactionRecordRequest request, String inventoryId, String id) {
        log.info("Updating transaction record for type: {}", request.getType());
        TransactionRecord transactionRecord = new TransactionRecord();
        transactionRecord.setId(id);
        transactionRecord.setType(request.getType());
        transactionRecord.setSource(request.getSource());
        transactionRecord.setDestination(request.getDestination());
        transactionRecord.setInventoryId(request.getInventoryId());
        transactionRecord.setInventoryMeasure(request.getInventoryMeasure());
        transactionRecord.setInventoryCategory(request.getInventoryCategory());
        transactionRecord.setInventoryUnit(request.getInventoryUnit());
        transactionRecord.setQuantity(request.getQuantity());
        transactionRecord.setStatus(request.getStatus());
        TransactionRecord savedRecord = transactionRecordRepository.update(transactionRecord);
        System.out.println(savedRecord.getStatus());
        // if (savedRecord.getStatus() == "Completed"){
            // Publish inventory status update to Kafka using the predefined schema
            try {
                var importRecord = inventoryProducer.createInventoryImportRecord(
                    savedRecord.getInventoryId(), 
                    savedRecord.getInventoryMeasure(),
                    savedRecord.getInventoryCategory(),
                    savedRecord.getInventoryUnit(),
                    savedRecord.getQuantity(),
                    savedRecord.getType()
                );
                inventoryProducer.sendInventoryStatusUpdate(savedRecord.getInventoryId(), importRecord);
                log.info("Successfully published inventory import event for ID: {}", savedRecord.getInventoryId());
            } catch (Exception e) {
                log.error("Failed to publish inventory status update for ID: {}, error: {}", 
                    savedRecord.getInventoryId(), e.getMessage());
                // Continue processing even if Kafka publishing fails
            }
        // }
        return mapToResponse(savedRecord);
    }
}
