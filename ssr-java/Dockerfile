FROM gradle:9.8.2-jdk27-alpine AS build
WORKDIR /app
COPY build.gradle ./
COPY src ./src
RUN gradle bootJar

FROM eclipse-temurin:27-jre-alpine-3.24
WORKDIR /app
COPY --from=build /app/build/libs/*.jar app.jar
RUN addgroup -S app && adduser -S -G app app
USER app
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=2s --start-period=40s CMD wget -q -O /dev/null http://127.0.0.1:8080/games || exit 1
ENTRYPOINT ["java", "-jar", "app.jar"]
