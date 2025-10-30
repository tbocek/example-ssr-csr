package ch.ost.i.dsl.di_ex5;

import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Service;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

interface NotificationService {
    void sendNotification(String to, String message);
}

@Service
@Profile("email")  // Active when spring.profiles.active=email
class EmailNotificationService_Profile implements NotificationService {
    @Override
    public void sendNotification(String to, String message) {
        System.out.println("📧 [EMAIL MODE] Email to " + to + ": " + message);
    }
}

@Service
@Profile("sms")  // Active when spring.profiles.active=sms
class SmsNotificationService_Profile implements NotificationService {
    @Override
    public void sendNotification(String to, String message) {
        System.out.println("📱 [SMS MODE] SMS to " + to + ": " + message);
    }
}

@RestController
@RequestMapping("/api/v5")
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
