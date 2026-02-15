# Runtime image - expects JAR to be built externally
FROM docker.io/amazoncorretto:25.0.2-alpine3.23

# Combine RUN commands to reduce layers and image size
RUN addgroup -S inventium && \
    adduser -S appuser -G inventium

WORKDIR /app

# Copy the pre-built JAR file from build directory
# Build the JAR externally with: ./gradlew build
COPY ./builder/libs/*.jar /app/application.jar

# Set ownership and permissions in a single layer
RUN chown appuser:inventium /app/application.jar && \
    chmod 555 /app && \
    chmod 444 /app/application.jar

# Run as non-privileged user
USER appuser

EXPOSE 14330

# Use exec form with optimized JVM flags for containers
ENTRYPOINT ["java", \
    "-XX:+UseContainerSupport", \
    "-XX:MaxRAMPercentage=75.0", \
    "-Djava.security.egd=file:/dev/./urandom", \
    "-jar", \
    "/app/application.jar"]
