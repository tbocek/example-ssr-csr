package ch.ost.i.dsl.di_ex3;

import org.springframework.stereotype.Service;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
@Service
class EmailService {
    public void sendNotification(String to, String message) {
        System.out.println("📧 Email to " + to + ": " + message);
    }
}

@RestController
@RequestMapping("/api/v3")
class UserController {
    
    private final EmailService emailService;
    
    // @Autowired is optional on constructor since Spring 4.3
    public UserController(EmailService emailService) {
        this.emailService = emailService;
    }
    
    @PostMapping("/register")
    public String registerUser(@RequestParam String username) {
        System.out.println("Registering user: " + username);
        emailService.sendNotification(username + "@example.com", "Welcome!");
        return "User registered: " + username;
    }
}