package com.inventium.inventory_transaction_service.config;

import com.inventium.inventory_transaction_service.entity.TransactionRecord;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbEnhancedClient;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbTable;
import software.amazon.awssdk.enhanced.dynamodb.TableSchema;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.ResourceNotFoundException;

import java.net.URI;

@Configuration
public class DynamoDBConfig {

    @Value("${aws.region:ap-southeast-7}")
    private String awsRegion;

    @Value("${aws.dynamodb.endpoint:}")
    private String dynamoDbEndpoint;

    @Bean
    public DynamoDbClient dynamoDbClient(){
        if (!dynamoDbEndpoint.isEmpty()){
            return DynamoDbClient.builder()
                .region(Region.of(awsRegion))
                .endpointOverride(URI.create(dynamoDbEndpoint))
                .build();
        }
        return DynamoDbClient.builder()
            .region(Region.of(awsRegion))
            .build();
    }

    @Bean
    public DynamoDbEnhancedClient dynamoDbEnhancedClient(DynamoDbClient dynamoDbClient){
        DynamoDbEnhancedClient dbEnhancedClient = DynamoDbEnhancedClient.builder()
            .dynamoDbClient(dynamoDbClient)
            .build();
        
        // Create table if it doesn't exist
        createTableIfNotExist(dbEnhancedClient);
        
        return dbEnhancedClient;
    }

    private void createTableIfNotExist(DynamoDbEnhancedClient enhancedDbEnhancedClient){
        try {
            DynamoDbTable<TransactionRecord> table = enhancedDbEnhancedClient.table("inventory-transaction-service", TableSchema.fromBean(TransactionRecord.class));
            table.describeTable();
            System.out.println("Table 'inventory-transaction-service' already exists");
        } catch (ResourceNotFoundException e){
            System.out.println("Creating table 'inventory-transaction-service'....");
            DynamoDbTable<TransactionRecord> table = enhancedDbEnhancedClient.table("inventory-transaction-service", TableSchema.fromBean(TransactionRecord.class));
            table.createTable();
            System.out.println("Table 'inventory-transaction-service' created successfully");
        } catch (Exception e) {
            System.err.println("Error checking/creating table: " + e.getMessage());
            e.printStackTrace();
        }
    }
}
