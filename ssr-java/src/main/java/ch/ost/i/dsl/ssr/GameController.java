package ch.ost.i.dsl.ssr;

import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;

@Controller
public class GameController {
    
    private final GameRepository gameRepository;
        
    public GameController(GameRepository gameRepository) {
        this.gameRepository = gameRepository;
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
        gameRepository.findById(id).ifPresent(game -> {
            game.addStar();
            gameRepository.save(game);
        });
        return "redirect:/games";
    }
}