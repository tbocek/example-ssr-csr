package ch.ost.i.dsl.di_ex1;

import org.springframework.web.bind.annotation.*;

// Plain classes - no Spring annotations
class EmailSender {
    public void send(String to, String message) {
        System.out.println("📧 Email to " + to + ": " + message);
    }
}

@RestController
@RequestMapping("/api/v1")
class UserController_Manual {
    
    private EmailSender emailSender;
    
    // Hard-coded dependency
    public UserController_Manual() {
        this.emailSender = new EmailSender(); 
    }
    
    @PostMapping("/register")
    public String registerUser(@RequestParam String username) {
        System.out.println("Registering user: " + username);
        emailSender.send(username + "@example.com", "Welcome!");
        return "User registered: " + username;
    }
}