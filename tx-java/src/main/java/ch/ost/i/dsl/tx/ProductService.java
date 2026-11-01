package ch.ost.i.dsl.tx;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class ProductService {
    
    private final ProductRepository productRepository;
    private final ProductStatisticsRepository statsRepository;
    
    public ProductService(ProductRepository productRepository, 
                      ProductStatisticsRepository statsRepository) {
        this.productRepository = productRepository;
        this.statsRepository = statsRepository;
    }
    
    /**
     * Without @Transactional: If buy() succeeds but updateStatistics() fails,
     * we get inconsistent data - the stock is updated but statistics are not.
     * 
     * With @Transactional: Either both operations succeed, or both are rolled back.
     * This ensures data consistency.
     */
    @Transactional
    public void buyWithStatistics(Long productId) {
        // Operation 1: Update the product
        Product product = productRepository.findByIdForUpdate(productId)
            .orElseThrow(() -> new IllegalArgumentException("Product not found"));
        if (product.getStock() < 1) {
            throw new IllegalStateException("Product is sold out");
        }
        product.buy();
        productRepository.save(product);
        
        //Network error
        //throw new RuntimeException("Simulated failure! Database connection lost!");
        
        // Operation 2: Update statistics
        ProductStatistics stats = statsRepository.findByProductId(productId)
            .orElse(new ProductStatistics(productId));
        stats.incrementTotalSold();
        stats.setLastUpdated(java.time.LocalDateTime.now());
        statsRepository.save(stats);
        
        // If any exception occurs here, BOTH operations are rolled back
        // Without @Transactional, the stock would drop but stats wouldn't update
    }
    
    /**
     * Example with explicit rollback on business logic violation
     */
    @Transactional
    public void transferStock(Long fromProductId, Long toProductId, int stockToTransfer) {
        if (stockToTransfer < 1) {
            throw new IllegalArgumentException("Stock to transfer must be positive");
        }
        
        productRepository.lockInIdOrder(java.util.List.of(fromProductId, toProductId));
        
        Product fromProduct = productRepository.findById(fromProductId)
            .orElseThrow(() -> new IllegalArgumentException("Source product not found"));
        Product toProduct = productRepository.findById(toProductId)
            .orElseThrow(() -> new IllegalArgumentException("Target product not found"));
        
        // Business rule validation
        if (fromProduct.getStock() < stockToTransfer) {
            throw new IllegalStateException("Not enough stock to transfer");
            // This exception triggers automatic rollback
        }
        
        // Both operations must succeed together
        fromProduct.setStock(fromProduct.getStock() - stockToTransfer);
        toProduct.setStock(toProduct.getStock() + stockToTransfer);
        
        productRepository.save(fromProduct);
        productRepository.save(toProduct);
        
        // Without @Transactional: fromProduct might be saved but toProduct save could fail
        // leaving the system in an inconsistent state
    }
    
    /**
     * Example showing rollback rules - only rollback on specific exceptions
     */
    @Transactional(rollbackFor = {IllegalStateException.class})
    public void conditionalRollback(Long productId) {
        // This will rollback on IllegalStateException
        // but NOT on other runtime exceptions (unless configured)
    }
}