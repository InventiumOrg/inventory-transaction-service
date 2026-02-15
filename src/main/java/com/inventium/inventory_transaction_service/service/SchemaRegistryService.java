package com.inventium.inventory_transaction_service.service;

import io.confluent.kafka.schemaregistry.client.CachedSchemaRegistryClient;
import io.confluent.kafka.schemaregistry.client.SchemaRegistryClient;
import io.confluent.kafka.schemaregistry.client.rest.exceptions.RestClientException;
import org.apache.avro.Schema;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import lombok.extern.slf4j.Slf4j;

import jakarta.annotation.PostConstruct;
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

@Service
@Slf4j
public class SchemaRegistryService {
    
    @Value("${kafka.inventory.schema.registry.url}")
    private String schemaRegistryUrl;
    
    @Value("${kafka.inventory.schema.registry.auth.user.info}")
    private String schemaRegistryAuthUserInfo;
    
    private SchemaRegistryClient schemaRegistryClient;
    private final Map<String, Schema> schemaCache = new HashMap<>();
    
    @PostConstruct
    public void initialize() {
        Map<String, String> configs = new HashMap<>();
        configs.put("basic.auth.credentials.source", "USER_INFO");
        configs.put("basic.auth.user.info", schemaRegistryAuthUserInfo);
        
        this.schemaRegistryClient = new CachedSchemaRegistryClient(schemaRegistryUrl, 100, configs);
        log.info("Schema Registry client initialized with URL: {}", schemaRegistryUrl);
    }
    
    public Schema getSchemaBySubject(String subject) {
        return getSchemaBySubject(subject, "latest");
    }
    
    public Schema getSchemaBySubject(String subject, String version) {
        String cacheKey = subject + ":" + version;
        
        if (schemaCache.containsKey(cacheKey)) {
            log.debug("Returning cached schema for subject: {} version: {}", subject, version);
            return schemaCache.get(cacheKey);
        }
        
        try {
            io.confluent.kafka.schemaregistry.client.SchemaMetadata schemaMetadata;
            
            if ("latest".equals(version)) {
                schemaMetadata = schemaRegistryClient.getLatestSchemaMetadata(subject);
            } else {
                int versionInt = Integer.parseInt(version);
                schemaMetadata = schemaRegistryClient.getSchemaMetadata(subject, versionInt);
            }
            
            Schema schema = new Schema.Parser().parse(schemaMetadata.getSchema());
            schemaCache.put(cacheKey, schema);
            
            log.info("Successfully fetched schema for subject: {} version: {} id: {}", 
                subject, schemaMetadata.getVersion(), schemaMetadata.getId());
            
            return schema;
            
        } catch (IOException | RestClientException e) {
            log.error("Failed to fetch schema for subject: {} version: {}", subject, version, e);
            throw new RuntimeException("Failed to fetch schema from registry", e);
        }
    }
    
    public Schema getSchemaById(int schemaId) {
        String cacheKey = "id:" + schemaId;
        
        if (schemaCache.containsKey(cacheKey)) {
            log.debug("Returning cached schema for ID: {}", schemaId);
            return schemaCache.get(cacheKey);
        }
        
        try {
            Schema schema = schemaRegistryClient.getById(schemaId);
            schemaCache.put(cacheKey, schema);
            
            log.info("Successfully fetched schema by ID: {}", schemaId);
            return schema;
            
        } catch (IOException | RestClientException e) {
            log.error("Failed to fetch schema by ID: {}", schemaId, e);
            throw new RuntimeException("Failed to fetch schema from registry", e);
        }
    }
    
    public void clearCache() {
        schemaCache.clear();
        log.info("Schema cache cleared");
    }
}