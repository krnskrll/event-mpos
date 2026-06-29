# Multi-stage build — no local Java/Gradle required, only Docker.
FROM gradle:8.7-jdk21 AS build
WORKDIR /home/gradle/src
COPY build.gradle.kts settings.gradle.kts ./
# Pre-resolve dependencies for better layer caching (best-effort).
RUN gradle dependencies --no-daemon > /dev/null 2>&1 || true
COPY src ./src
RUN gradle bootJar --no-daemon

FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=build /home/gradle/src/build/libs/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
