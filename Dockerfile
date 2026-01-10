FROM openjdk:25-jdk-slim

WORKDIR /app

COPY build/libs/*.jar app.jar

EXPOSE 14330

ENTRYPOINT ["java", "-jar", "app.jar"]