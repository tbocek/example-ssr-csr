package ch.ost.i.dsl.ssr;

import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.PostMapping;

@Controller
public class UserController {
    
    private final UserRepository userRepository;
        
    public UserController(UserRepository userRepository) {
        this.userRepository = userRepository;
    }
        
    @GetMapping({"/", "/users"})
    public String listUsers(Model model) {
        model.addAttribute("users", userRepository.findAll());
        model.addAttribute("newUser", new User());
        return "users";
    }
    
    @PostMapping("/users")
    public String createUser(@ModelAttribute User user) {
        userRepository.save(user);
        return "redirect:/users";
    }
}