package ch.ost.i.dsl.tx;

import jakarta.persistence.*;
import java.time.LocalDateTime;

@Entity
@Table(name = "product_statistics")
public class ProductStatistics {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(unique = true)
    private Long productId;
    
    private Integer totalSold = 0;
    private LocalDateTime lastUpdated;
    
    public ProductStatistics() {}
    
    public ProductStatistics(Long productId) {
        this.productId = productId;
        this.totalSold = 0;
        this.lastUpdated = LocalDateTime.now();
    }
    
    public void incrementTotalSold() {
        this.totalSold++;
    }
    
    // Getters and setters
    public Long getId() { return id; }
    public void setId(Long id) { this.id = id; }
    
    public Long getProductId() { return productId; }
    public void setProductId(Long productId) { this.productId = productId; }
    
    public Integer getTotalSold() { return totalSold; }
    public void setTotalSold(Integer totalSold) { this.totalSold = totalSold; }
    
    public LocalDateTime getLastUpdated() { return lastUpdated; }
    public void setLastUpdated(LocalDateTime lastUpdated) { this.lastUpdated = lastUpdated; }
}