package ch.ost.i.dsl.tx;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import java.util.Optional;

@Repository
public interface ProductStatisticsRepository extends JpaRepository<ProductStatistics, Long> {
    Optional<ProductStatistics> findByProductId(Long productId);
}