package ch.ost.i.dsl.di_ex2;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

class EmailService {

    public void sendNotification(String to, String message) {
        System.out.println("📧 Email to " + to + ": " + message);
    }
}

@Configuration
class AppConfig {

    @Bean
    public EmailService emailService() {
        return new EmailService(); // YOU create it
    }

    @Bean
    public UserController userController() {
        return new UserController(emailService()); // YOU wire it
    }
}

@RestController
@RequestMapping("/api/v2")
class UserController {

    private EmailService emailService;

    // Plain constructor - no @Autowired
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
