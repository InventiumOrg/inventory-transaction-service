# -----------------------------------------------------------------------------
# Stage 1: compile with Gradle wrapper (JDK 25, matches build.gradle toolchain)
# -----------------------------------------------------------------------------
FROM docker.io/amazoncorretto:25.0.2-alpine3.23 AS builder

WORKDIR /workspace

# Gradle wrapper + build definition (good layer cache when only src changes)
COPY gradle/ gradle/
COPY gradlew settings.gradle build.gradle ./

RUN chmod +x gradlew \
	&& sed -i 's/\r$//' gradlew

# Warm dependency cache without compiling app code yet
RUN ./gradlew --no-daemon dependencies

COPY src/ src/

RUN ./gradlew --no-daemon bootJar -x test \
	&& JAR="$(find build/libs -maxdepth 1 -name '*.jar' ! -name '*-plain.jar' | head -n 1)" \
	&& test -n "$JAR" \
	&& cp "$JAR" /workspace/application.jar

# -----------------------------------------------------------------------------
# Stage 2: minimal runtime (JRE only)
# -----------------------------------------------------------------------------
FROM docker.io/amazoncorretto:25.0.2-alpine3.23

RUN addgroup -S inventium \
	&& adduser -S appuser -G inventium

WORKDIR /app

COPY --from=builder /workspace/application.jar /app/application.jar

RUN chown appuser:inventium /app/application.jar \
	&& chmod 555 /app \
	&& chmod 444 /app/application.jar

USER appuser

EXPOSE 14330

ENTRYPOINT ["java", \
	"-XX:+UseContainerSupport", \
	"-XX:MaxRAMPercentage=75.0", \
	"-Djava.security.egd=file:/dev/./urandom", \
	"-jar", \
	"/app/application.jar"]
