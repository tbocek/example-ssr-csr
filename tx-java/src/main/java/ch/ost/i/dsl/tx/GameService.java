package ch.ost.i.dsl.tx;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class GameService {
    
    private final GameRepository gameRepository;
    private final GameStatisticsRepository statsRepository;
    
    public GameService(GameRepository gameRepository, 
                      GameStatisticsRepository statsRepository) {
        this.gameRepository = gameRepository;
        this.statsRepository = statsRepository;
    }
    
    /**
     * Without @Transactional: If addStar() succeeds but updateStatistics() fails,
     * we get inconsistent data - the star count is updated but statistics are not.
     * 
     * With @Transactional: Either both operations succeed, or both are rolled back.
     * This ensures data consistency.
     */
    @Transactional
    public void addStarWithStatistics(Long gameId) {
        // Operation 1: Update the game
        Game game = gameRepository.findById(gameId)
            .orElseThrow(() -> new IllegalArgumentException("Game not found"));
        game.addStar();
        gameRepository.save(game);
        
        //Network error
        //throw new RuntimeException("Simulated failure! Database connection lost!");
        
        // Operation 2: Update statistics
        GameStatistics stats = statsRepository.findByGameId(gameId)
            .orElse(new GameStatistics(gameId));
        stats.incrementTotalStars();
        stats.setLastUpdated(java.time.LocalDateTime.now());
        statsRepository.save(stats);
        
        // If any exception occurs here, BOTH operations are rolled back
        // Without @Transactional, the game would have the star but stats wouldn't update
    }
    
    /**
     * Example with explicit rollback on business logic violation
     */
    @Transactional
    public void transferStars(Long fromGameId, Long toGameId, int starsToTransfer) {
        Game fromGame = gameRepository.findById(fromGameId)
            .orElseThrow(() -> new IllegalArgumentException("Source game not found"));
        Game toGame = gameRepository.findById(toGameId)
            .orElseThrow(() -> new IllegalArgumentException("Target game not found"));
        
        // Business rule validation
        if (fromGame.getStars() < starsToTransfer) {
            throw new IllegalStateException("Not enough stars to transfer");
            // This exception triggers automatic rollback
        }
        
        // Both operations must succeed together
        fromGame.setStars(fromGame.getStars() - starsToTransfer);
        toGame.setStars(toGame.getStars() + starsToTransfer);
        
        gameRepository.save(fromGame);
        gameRepository.save(toGame);
        
        // Without @Transactional: fromGame might be saved but toGame save could fail
        // leaving the system in an inconsistent state
    }
    
    /**
     * Example showing rollback rules - only rollback on specific exceptions
     */
    @Transactional(rollbackFor = {IllegalStateException.class})
    public void conditionalRollback(Long gameId) {
        // This will rollback on IllegalStateException
        // but NOT on other runtime exceptions (unless configured)
    }
}