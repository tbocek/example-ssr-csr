package ch.ost.i.dsl.di_ex4;

import org.springframework.stereotype.Service;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

interface NotificationService {
    void sendNotification(String to, String message);
}

// Email implementation
@Service
class EmailNotificationService implements NotificationService {

    @Override
    public void sendNotification(String to, String message) {
        System.out.println("📧 Email to " + to + ": " + message);
    }
}

//@Service  // <-- Uncomment this to use SMS instead of Email
class SmsNotificationService implements NotificationService {

    @Override
    public void sendNotification(String to, String message) {
        System.out.println("📱 SMS to " + to + ": " + message);
    }
}

@RestController
@RequestMapping("/api/v4")
class UserController {

    private final NotificationService notificationService;

    // @Autowired is optional on constructor since Spring 4.3
    public UserController(NotificationService notificationService) {
        this.notificationService = notificationService;
    }

    @PostMapping("/register")
    public String registerUser(@RequestParam String username) {
        System.out.println("Registering user: " + username);
        notificationService.sendNotification(username + "@example.com", "Welcome!");
        return "User registered: " + username;
    }
}
