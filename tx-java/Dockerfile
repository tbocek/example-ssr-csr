FROM gradle:9.8.2-jdk27-alpine AS build
WORKDIR /app
COPY build.gradle ./
COPY src ./src
RUN gradle bootJar

FROM eclipse-temurin:27-jre-alpine-3.24
WORKDIR /app
COPY --from=build /app/build/libs/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]