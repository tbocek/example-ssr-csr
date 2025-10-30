FROM gradle:jdk25-alpine AS build
WORKDIR /app
COPY build.gradle ./
COPY src ./src
RUN gradle bootJar

FROM eclipse-temurin:25-jre-alpine
WORKDIR /app
COPY --from=build /app/build/libs/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]

# run with: docker build . -t ex2 && docker run -p 8080:8080 ex2
# test with: curl -X POST http://localhost:8080/api/v2/register -d "username=john" 