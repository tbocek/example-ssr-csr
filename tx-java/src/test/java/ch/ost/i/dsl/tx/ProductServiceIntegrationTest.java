package ch.ost.i.dsl.tx;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.springframework.transaction.annotation.Transactional;
import org.testcontainers.postgresql.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@Testcontainers
class ProductServiceIntegrationTest {

    @Container
    static PostgreSQLContainer postgres = new PostgreSQLContainer("postgres:18-alpine")
            .withDatabaseName("testdb")
            .withUsername("test")
            .withPassword("test");

    @DynamicPropertySource
    static void configureProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
    }

    @Autowired
    private ProductService productService;

    @Autowired
    private ProductRepository productRepository;

    @Autowired
    private ProductStatisticsRepository statsRepository;

    @Test
    @Transactional
    void transferStock_WithTransaction_RollsBackOnError() {
        // Arrange
        Product fromProduct = new Product("Source Product", "Has stock");
        fromProduct.setStock(10);
        fromProduct = productRepository.save(fromProduct);

        Product toProduct = new Product("Target Product", "Needs stock");
        toProduct.setStock(5);
        toProduct = productRepository.save(toProduct);

        Long fromId = fromProduct.getId();
        Long toId = toProduct.getId();

        // Act - try to transfer more stock than available
        assertThrows(IllegalStateException.class, () -> {
            productService.transferStock(fromId, toId, 15);
        });

        // Assert - both products should be unchanged due to rollback
        Product fromProductAfter = productRepository.findById(fromId).orElseThrow();
        Product toProductAfter = productRepository.findById(toId).orElseThrow();

        assertEquals(10, fromProductAfter.getStock(), "Source product stock should be unchanged after rollback");
        assertEquals(5, toProductAfter.getStock(), "Target product stock should be unchanged after rollback");
    }

    @Test
    @Transactional
    void transferStock_WithTransaction_CommitsOnSuccess() {
        // Arrange
        Product fromProduct = new Product("Source Product", "Has stock");
        fromProduct.setStock(10);
        fromProduct = productRepository.save(fromProduct);

        Product toProduct = new Product("Target Product", "Needs stock");
        toProduct.setStock(5);
        toProduct = productRepository.save(toProduct);

        Long fromId = fromProduct.getId();
        Long toId = toProduct.getId();

        // Act
        productService.transferStock(fromId, toId, 3);

        // Assert - transfer should succeed
        Product fromProductAfter = productRepository.findById(fromId).orElseThrow();
        Product toProductAfter = productRepository.findById(toId).orElseThrow();

        assertEquals(7, fromProductAfter.getStock(), "Source product should have 3 less stock");
        assertEquals(8, toProductAfter.getStock(), "Target product should have 3 more stock");
    }

    @Test
    @Transactional
    void buyWithStatistics_CreatesStatisticsIfNotExists() {
        // Arrange
        Product product = new Product("New Product", "No stats yet");
        product.setStock(1);
        product = productRepository.save(product);
        Long productId = product.getId();

        // Act
        productService.buyWithStatistics(productId);

        // Assert
        Product updatedProduct = productRepository.findById(productId).orElseThrow();
        assertEquals(0, updatedProduct.getStock());

        ProductStatistics stats = statsRepository.findByProductId(productId).orElseThrow();
        assertEquals(1, stats.getTotalSold());
        assertNotNull(stats.getLastUpdated());
    }

    @Test
    @Transactional
    void buyWithStatistics_UpdatesExistingStatistics() {
        // Arrange
        Product product = new Product("Existing Product", "Has stats");
        product.setStock(5);
        product = productRepository.save(product);
        Long productId = product.getId();

        ProductStatistics stats = new ProductStatistics(productId);
        stats.setTotalSold(5);
        statsRepository.save(stats);

        // Act
        productService.buyWithStatistics(productId);

        // Assert
        Product updatedProduct = productRepository.findById(productId).orElseThrow();
        assertEquals(4, updatedProduct.getStock());

        ProductStatistics updatedStats = statsRepository.findByProductId(productId).orElseThrow();
        assertEquals(6, updatedStats.getTotalSold());
    }

    @Test
    @Transactional
    void transferStock_EdgeCase_TransferAllStock() {
        // Arrange
        Product fromProduct = new Product("Source", "Transfer all");
        fromProduct.setStock(5);
        fromProduct = productRepository.save(fromProduct);

        Product toProduct = new Product("Target", "Receive all");
        toProduct.setStock(0);
        toProduct = productRepository.save(toProduct);

        Long fromId = fromProduct.getId();
        Long toId = toProduct.getId();

        // Act
        productService.transferStock(fromId, toId, 5);

        // Assert
        Product fromProductAfter = productRepository.findById(fromId).orElseThrow();
        Product toProductAfter = productRepository.findById(toId).orElseThrow();

        assertEquals(0, fromProductAfter.getStock());
        assertEquals(5, toProductAfter.getStock());
    }
}