package ch.ost.i.dsl.tx;

import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;

@Controller
public class GameController {
    
    private final GameRepository gameRepository;
    private final GameService gameService;  // Use service layer for transactional operations
        
    public GameController(GameRepository gameRepository, GameService gameService) {
        this.gameRepository = gameRepository;
        this.gameService = gameService;
    }
        
    @GetMapping({"/", "/games"})
    public String listGames(Model model) {
        model.addAttribute("games", gameRepository.findAll());
        model.addAttribute("newGame", new Game());
        return "games";
    }
    
    @PostMapping("/games")
    public String createGame(@ModelAttribute Game game) {
        gameRepository.save(game);
        return "redirect:/games";
    }
    
    @PostMapping("/games/{id}/star")
    public String addStar(@PathVariable Long id) {
        try {
            gameService.addStarWithStatistics(id);
        } catch (IllegalArgumentException e) {
            // Handle game not found
            return "redirect:/games?error=not-found";
        }
        return "redirect:/games";
    }
    
    @PostMapping("/games/{fromId}/transfer/{toId}")
    public String transferStars(@PathVariable Long fromId, 
                               @PathVariable Long toId,
                               @ModelAttribute("stars") int stars) {
        try {
            gameService.transferStars(fromId, toId, stars);
        } catch (IllegalStateException e) {
            return "redirect:/games?error=insufficient-stars";
        }
        return "redirect:/games";
    }
}