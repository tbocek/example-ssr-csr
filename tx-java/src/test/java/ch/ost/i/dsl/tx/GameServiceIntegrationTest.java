package ch.ost.i.dsl.tx;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.springframework.transaction.annotation.Transactional;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@Testcontainers
class GameServiceIntegrationTest {

    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:18-alpine")
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
    private GameService gameService;

    @Autowired
    private GameRepository gameRepository;

    @Autowired
    private GameStatisticsRepository statsRepository;

    @Test
    @Transactional
    void transferStars_WithTransaction_RollsBackOnError() {
        // Arrange
        Game fromGame = new Game("Source Game", "Has stars");
        fromGame.setStars(10);
        fromGame = gameRepository.save(fromGame);

        Game toGame = new Game("Target Game", "Needs stars");
        toGame.setStars(5);
        toGame = gameRepository.save(toGame);

        Long fromId = fromGame.getId();
        Long toId = toGame.getId();

        // Act - try to transfer more stars than available
        assertThrows(IllegalStateException.class, () -> {
            gameService.transferStars(fromId, toId, 15);
        });

        // Assert - both games should be unchanged due to rollback
        Game fromGameAfter = gameRepository.findById(fromId).orElseThrow();
        Game toGameAfter = gameRepository.findById(toId).orElseThrow();

        assertEquals(10, fromGameAfter.getStars(), "Source game stars should be unchanged after rollback");
        assertEquals(5, toGameAfter.getStars(), "Target game stars should be unchanged after rollback");
    }

    @Test
    @Transactional
    void transferStars_WithTransaction_CommitsOnSuccess() {
        // Arrange
        Game fromGame = new Game("Source Game", "Has stars");
        fromGame.setStars(10);
        fromGame = gameRepository.save(fromGame);

        Game toGame = new Game("Target Game", "Needs stars");
        toGame.setStars(5);
        toGame = gameRepository.save(toGame);

        Long fromId = fromGame.getId();
        Long toId = toGame.getId();

        // Act
        gameService.transferStars(fromId, toId, 3);

        // Assert - transfer should succeed
        Game fromGameAfter = gameRepository.findById(fromId).orElseThrow();
        Game toGameAfter = gameRepository.findById(toId).orElseThrow();

        assertEquals(7, fromGameAfter.getStars(), "Source game should have 3 less stars");
        assertEquals(8, toGameAfter.getStars(), "Target game should have 3 more stars");
    }

    @Test
    @Transactional
    void addStarWithStatistics_CreatesStatisticsIfNotExists() {
        // Arrange
        Game game = new Game("New Game", "No stats yet");
        game.setStars(0);
        game = gameRepository.save(game);
        Long gameId = game.getId();

        // Act
        gameService.addStarWithStatistics(gameId);

        // Assert
        Game updatedGame = gameRepository.findById(gameId).orElseThrow();
        assertEquals(1, updatedGame.getStars());

        GameStatistics stats = statsRepository.findByGameId(gameId).orElseThrow();
        assertEquals(1, stats.getTotalStars());
        assertNotNull(stats.getLastUpdated());
    }

    @Test
    @Transactional
    void addStarWithStatistics_UpdatesExistingStatistics() {
        // Arrange
        Game game = new Game("Existing Game", "Has stats");
        game.setStars(5);
        game = gameRepository.save(game);
        Long gameId = game.getId();

        GameStatistics stats = new GameStatistics(gameId);
        stats.setTotalStars(5);
        statsRepository.save(stats);

        // Act
        gameService.addStarWithStatistics(gameId);

        // Assert
        Game updatedGame = gameRepository.findById(gameId).orElseThrow();
        assertEquals(6, updatedGame.getStars());

        GameStatistics updatedStats = statsRepository.findByGameId(gameId).orElseThrow();
        assertEquals(6, updatedStats.getTotalStars());
    }

    @Test
    @Transactional
    void transferStars_EdgeCase_TransferAllStars() {
        // Arrange
        Game fromGame = new Game("Source", "Transfer all");
        fromGame.setStars(5);
        fromGame = gameRepository.save(fromGame);

        Game toGame = new Game("Target", "Receive all");
        toGame.setStars(0);
        toGame = gameRepository.save(toGame);

        Long fromId = fromGame.getId();
        Long toId = toGame.getId();

        // Act
        gameService.transferStars(fromId, toId, 5);

        // Assert
        Game fromGameAfter = gameRepository.findById(fromId).orElseThrow();
        Game toGameAfter = gameRepository.findById(toId).orElseThrow();

        assertEquals(0, fromGameAfter.getStars());
        assertEquals(5, toGameAfter.getStars());
    }
}