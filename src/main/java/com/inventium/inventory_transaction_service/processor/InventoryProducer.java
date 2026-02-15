package com.inventium.inventory_transaction_service.processor;

import java.util.Properties;
import org.apache.kafka.clients.producer.KafkaProducer;
import org.apache.kafka.clients.producer.ProducerRecord;
import org.apache.kafka.common.serialization.StringSerializer;
import org.apache.avro.Schema;
import org.apache.avro.generic.GenericData;
import org.apache.avro.generic.GenericRecord;
import io.confluent.kafka.serializers.KafkaAvroSerializer;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import lombok.extern.slf4j.Slf4j;
import com.inventium.inventory_transaction_service.service.SchemaRegistryService;

@Component
@Slf4j
public class InventoryProducer {
    
    @Autowired
    private SchemaRegistryService schemaRegistryService;
    
    @Value("${kafka.inventory.topic.name}")
    private String topicName;
    
    @Value("${kafka.inventory.ssl.ca.pem.location}")
    private String caPemLocation;
    
    @Value("${kafka.inventory.ssl.svc.pem.location}")
    private String svcPemLocation;
    
    @Value("${kafka.inventory.bootstrap.servers}")
    private String bootstrapServers;
    
    @Value("${kafka.inventory.schema.registry.url}")
    private String schemaRegistryUrl;
    
    @Value("${kafka.inventory.schema.registry.auth.user.info}")
    private String schemaRegistryAuthUserInfo;
    
    @Value("${kafka.inventory.auto.register.schemas:true}")
    private String autoRegisterSchemas;
    
    @Value("${kafka.inventory.schema.subject}")
    private String schemaSubject;
    
    private KafkaProducer<String, GenericRecord> producer;
    
    
    public void initializeProducer() {
        if (producer == null) {
            Properties props = new Properties();
            props.setProperty("auto.register.schemas", autoRegisterSchemas);
            props.setProperty("bootstrap.servers", bootstrapServers);
            props.setProperty("schema.registry.url", schemaRegistryUrl);
            props.setProperty("basic.auth.credentials.source", "USER_INFO");
            props.setProperty("basic.auth.user.info", schemaRegistryAuthUserInfo);
            props.setProperty("security.protocol", "SSL");
            props.setProperty("ssl.truststore.location", caPemLocation);
            props.setProperty("ssl.keystore.type", "PEM");
            props.setProperty("ssl.truststore.type", "PEM");
            props.setProperty("ssl.keystore.location", svcPemLocation);
            props.setProperty("key.serializer", StringSerializer.class.getName());
            props.setProperty("value.serializer", KafkaAvroSerializer.class.getName());
            
            producer = new KafkaProducer<>(props);
            log.info("Inventory Kafka producer initialized for topic: {}", topicName);
        }
    }
    
    public void sendInventoryStatusUpdate(String key, GenericRecord record) {
        try {
            initializeProducer();
            ProducerRecord<String, GenericRecord> producerRecord = 
                new ProducerRecord<>(topicName, key, record);
            
            producer.send(producerRecord, (metadata, exception) -> {
                if (exception != null) {
                    log.error("Failed to send inventory status update: {}", exception.getMessage());
                } else {
                    log.info("Inventory status update sent successfully to topic: {} at offset: {}", 
                        metadata.topic(), metadata.offset());
                }
            });
        } catch (Exception e) {
            log.error("Error sending inventory status update: {}", e.getMessage(), e);
        }
    }
    
    public GenericRecord createInventoryImportRecord(String inventoryId, String inventoryMeasure, 
                                                    String inventoryCategory, String inventoryUnit, int quantity, String type) {
        try {
            Schema schema = schemaRegistryService.getSchemaBySubject(schemaSubject);
            GenericRecord record = new GenericData.Record(schema);
            record.put("inventoryId", inventoryId);
            record.put("inventoryMeasure", inventoryMeasure);
            record.put("inventoryCategory", inventoryCategory);
            record.put("inventoryUnit", inventoryUnit);
            record.put("quantity", quantity);
            record.put("type", type);
            
            log.debug("Created inventory import record for inventoryId: {}", inventoryId);
            return record;
        } catch (Exception e) {
            log.error("Error creating inventory import record: {}", e.getMessage(), e);
            throw new RuntimeException("Failed to create inventory import record", e);
        }
    }
    
    public GenericRecord createInventoryStatusRecord(String inventoryId, String status, String transactionId) {
        try {
            Schema schema = schemaRegistryService.getSchemaBySubject(schemaSubject);
            GenericRecord record = new GenericData.Record(schema);
            record.put("inventoryId", inventoryId);
            record.put("status", status);
            record.put("transactionId", transactionId);
            record.put("timestamp", System.currentTimeMillis());
            
            log.debug("Created inventory status record for inventoryId: {} with status: {}", inventoryId, status);
            return record;
        } catch (Exception e) {
            log.error("Error creating inventory status record: {}", e.getMessage(), e);
            throw new RuntimeException("Failed to create inventory status record", e);
        }
    }
    
    public void close() {
        if (producer != null) {
            producer.close();
            log.info("Inventory Kafka producer closed");
        }
    }
}