package ch.ost.i.dsl.tx;

import jakarta.persistence.*;
import java.time.LocalDateTime;

@Entity
@Table(name = "game_statistics")
public class GameStatistics {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(unique = true)
    private Long gameId;
    
    private Integer totalStars = 0;
    private LocalDateTime lastUpdated;
    
    public GameStatistics() {}
    
    public GameStatistics(Long gameId) {
        this.gameId = gameId;
        this.totalStars = 0;
        this.lastUpdated = LocalDateTime.now();
    }
    
    public void incrementTotalStars() {
        this.totalStars++;
    }
    
    // Getters and setters
    public Long getId() { return id; }
    public void setId(Long id) { this.id = id; }
    
    public Long getGameId() { return gameId; }
    public void setGameId(Long gameId) { this.gameId = gameId; }
    
    public Integer getTotalStars() { return totalStars; }
    public void setTotalStars(Integer totalStars) { this.totalStars = totalStars; }
    
    public LocalDateTime getLastUpdated() { return lastUpdated; }
    public void setLastUpdated(LocalDateTime lastUpdated) { this.lastUpdated = lastUpdated; }
}