package com.inventium.inventory_transaction_service.repository;

import com.inventium.inventory_transaction_service.entity.TransactionRecord;
import org.springframework.stereotype.Repository;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbEnhancedClient;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbTable;
import software.amazon.awssdk.enhanced.dynamodb.TableSchema;

import java.util.Optional;

@Repository
public class TransactionRecordRepository {
    private final DynamoDbTable<TransactionRecord> table;

    public TransactionRecordRepository(DynamoDbEnhancedClient dynamoDbEnhancedClient){
        this.table = dynamoDbEnhancedClient.table("inventory-transaction-service", TableSchema.fromBean(TransactionRecord.class));
    }

    public TransactionRecord save(TransactionRecord transactionRecord){
        table.putItem(transactionRecord);
        return transactionRecord;
    }

    public TransactionRecord findById(String inventoryId, String id) {
        return table.getItem(r -> r.key(k -> k.partitionValue(inventoryId).sortValue(id)));
    }

    public TransactionRecord update(TransactionRecord transactionRecord){
        table.updateItem(transactionRecord);
        return transactionRecord;
    }
    
    public Optional<TransactionRecord> findByIdOptional(String inventoryId, String id) {
        TransactionRecord record = findById(inventoryId, id);
        return Optional.ofNullable(record);
    }
}
