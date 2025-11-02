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
class GameServiceTest {

    @Mock
    private GameRepository gameRepository;

    @Mock
    private GameStatisticsRepository statsRepository;

    @InjectMocks
    private GameService gameService;

    private Game testGame;
    private GameStatistics testStats;

    @BeforeEach
    void setUp() {
        testGame = new Game("Test Game", "Test Description");
        testGame.setId(1L);
        testGame.setStars(5);

        testStats = new GameStatistics(1L);
        testStats.setTotalStars(5);
    }

    @Test
    void addStarWithStatistics_Success() {
        // Arrange
        when(gameRepository.findById(1L)).thenReturn(Optional.of(testGame));
        when(statsRepository.findByGameId(1L)).thenReturn(Optional.of(testStats));
        when(gameRepository.save(any(Game.class))).thenReturn(testGame);
        when(statsRepository.save(any(GameStatistics.class))).thenReturn(testStats);

        // Act
        gameService.addStarWithStatistics(1L);

        // Assert
        assertEquals(6, testGame.getStars());
        verify(gameRepository).save(testGame);
        verify(statsRepository).save(testStats);
    }

    @Test
    void addStarWithStatistics_GameNotFound_ThrowsException() {
        // Arrange
        when(gameRepository.findById(999L)).thenReturn(Optional.empty());

        // Act & Assert
        IllegalArgumentException exception = assertThrows(
            IllegalArgumentException.class,
            () -> gameService.addStarWithStatistics(999L)
        );
        
        assertEquals("Game not found", exception.getMessage());
        verify(gameRepository, never()).save(any());
        verify(statsRepository, never()).save(any());
    }

    @Test
    void transferStars_Success() {
        // Arrange
        Game fromGame = new Game("Game 1", "Description 1");
        fromGame.setId(1L);
        fromGame.setStars(10);

        Game toGame = new Game("Game 2", "Description 2");
        toGame.setId(2L);
        toGame.setStars(5);

        when(gameRepository.findById(1L)).thenReturn(Optional.of(fromGame));
        when(gameRepository.findById(2L)).thenReturn(Optional.of(toGame));
        when(gameRepository.save(any(Game.class))).thenAnswer(i -> i.getArguments()[0]);

        // Act
        gameService.transferStars(1L, 2L, 3);

        // Assert
        assertEquals(7, fromGame.getStars());
        assertEquals(8, toGame.getStars());
        verify(gameRepository, times(2)).save(any(Game.class));
    }

    @Test
    void transferStars_InsufficientStars_ThrowsException() {
        // Arrange
        Game fromGame = new Game("Game 1", "Description 1");
        fromGame.setId(1L);
        fromGame.setStars(2);

        Game toGame = new Game("Game 2", "Description 2");
        toGame.setId(2L);
        toGame.setStars(5);

        when(gameRepository.findById(1L)).thenReturn(Optional.of(fromGame));
        when(gameRepository.findById(2L)).thenReturn(Optional.of(toGame));

        // Act & Assert
        IllegalStateException exception = assertThrows(
            IllegalStateException.class,
            () -> gameService.transferStars(1L, 2L, 5)
        );
        
        assertEquals("Not enough stars to transfer", exception.getMessage());
        verify(gameRepository, never()).save(any());
    }

    @Test
    void transferStars_SourceGameNotFound_ThrowsException() {
        // Arrange
        when(gameRepository.findById(999L)).thenReturn(Optional.empty());

        // Act & Assert
        IllegalArgumentException exception = assertThrows(
            IllegalArgumentException.class,
            () -> gameService.transferStars(999L, 2L, 3)
        );
        
        assertEquals("Source game not found", exception.getMessage());
    }

    @Test
    void transferStars_TargetGameNotFound_ThrowsException() {
        // Arrange
        Game fromGame = new Game("Game 1", "Description 1");
        fromGame.setId(1L);
        fromGame.setStars(10);

        when(gameRepository.findById(1L)).thenReturn(Optional.of(fromGame));
        when(gameRepository.findById(999L)).thenReturn(Optional.empty());

        // Act & Assert
        IllegalArgumentException exception = assertThrows(
            IllegalArgumentException.class,
            () -> gameService.transferStars(1L, 999L, 3)
        );
        
        assertEquals("Target game not found", exception.getMessage());
    }
}