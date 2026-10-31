package ch.ost.i.dsl.tx;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class ProductServiceTest {

    @Mock
    private ProductRepository productRepository;

    @Mock
    private ProductStatisticsRepository statsRepository;

    @InjectMocks
    private ProductService productService;

    private Product testProduct;
    private ProductStatistics testStats;

    @BeforeEach
    void setUp() {
        testProduct = new Product("Test Product", "Test Description");
        testProduct.setId(1L);
        testProduct.setStock(5);

        testStats = new ProductStatistics(1L);
        testStats.setTotalSold(5);
    }

    @Test
    void buyWithStatistics_Success() {
        // Arrange
        when(productRepository.findById(1L)).thenReturn(Optional.of(testProduct));
        when(statsRepository.findByProductId(1L)).thenReturn(Optional.of(testStats));
        when(productRepository.save(any(Product.class))).thenReturn(testProduct);
        when(statsRepository.save(any(ProductStatistics.class))).thenReturn(testStats);

        // Act
        productService.buyWithStatistics(1L);

        // Assert
        assertEquals(4, testProduct.getStock());
        assertEquals(6, testStats.getTotalSold());
        verify(productRepository).save(testProduct);
        verify(statsRepository).save(testStats);
    }

    @Test
    void buyWithStatistics_ProductNotFound_ThrowsException() {
        // Arrange
        when(productRepository.findById(999L)).thenReturn(Optional.empty());

        // Act & Assert
        IllegalArgumentException exception = assertThrows(
            IllegalArgumentException.class,
            () -> productService.buyWithStatistics(999L)
        );
        
        assertEquals("Product not found", exception.getMessage());
        verify(productRepository, never()).save(any());
        verify(statsRepository, never()).save(any());
    }

    @Test
    void buyWithStatistics_SoldOut_ThrowsException() {
        // Arrange
        testProduct.setStock(0);
        when(productRepository.findById(1L)).thenReturn(Optional.of(testProduct));

        // Act & Assert
        assertThrows(IllegalStateException.class, () -> productService.buyWithStatistics(1L));
        verify(productRepository, never()).save(any());
        verify(statsRepository, never()).save(any());
    }

    @Test
    void transferStock_Success() {
        // Arrange
        Product fromProduct = new Product("Product 1", "Description 1");
        fromProduct.setId(1L);
        fromProduct.setStock(10);

        Product toProduct = new Product("Product 2", "Description 2");
        toProduct.setId(2L);
        toProduct.setStock(5);

        when(productRepository.findById(1L)).thenReturn(Optional.of(fromProduct));
        when(productRepository.findById(2L)).thenReturn(Optional.of(toProduct));
        when(productRepository.save(any(Product.class))).thenAnswer(i -> i.getArguments()[0]);

        // Act
        productService.transferStock(1L, 2L, 3);

        // Assert
        assertEquals(7, fromProduct.getStock());
        assertEquals(8, toProduct.getStock());
        verify(productRepository, times(2)).save(any(Product.class));
    }

    @Test
    void transferStock_InsufficientStock_ThrowsException() {
        // Arrange
        Product fromProduct = new Product("Product 1", "Description 1");
        fromProduct.setId(1L);
        fromProduct.setStock(2);

        Product toProduct = new Product("Product 2", "Description 2");
        toProduct.setId(2L);
        toProduct.setStock(5);

        when(productRepository.findById(1L)).thenReturn(Optional.of(fromProduct));
        when(productRepository.findById(2L)).thenReturn(Optional.of(toProduct));

        // Act & Assert
        IllegalStateException exception = assertThrows(
            IllegalStateException.class,
            () -> productService.transferStock(1L, 2L, 5)
        );
        
        assertEquals("Not enough stock to transfer", exception.getMessage());
        verify(productRepository, never()).save(any());
    }

    @Test
    void transferStock_SourceProductNotFound_ThrowsException() {
        // Arrange
        when(productRepository.findById(999L)).thenReturn(Optional.empty());

        // Act & Assert
        IllegalArgumentException exception = assertThrows(
            IllegalArgumentException.class,
            () -> productService.transferStock(999L, 2L, 3)
        );
        
        assertEquals("Source product not found", exception.getMessage());
    }

    @Test
    void transferStock_TargetProductNotFound_ThrowsException() {
        // Arrange
        Product fromProduct = new Product("Product 1", "Description 1");
        fromProduct.setId(1L);
        fromProduct.setStock(10);

        when(productRepository.findById(1L)).thenReturn(Optional.of(fromProduct));
        when(productRepository.findById(999L)).thenReturn(Optional.empty());

        // Act & Assert
        IllegalArgumentException exception = assertThrows(
            IllegalArgumentException.class,
            () -> productService.transferStock(1L, 999L, 3)
        );
        
        assertEquals("Target product not found", exception.getMessage());
    }
}