package ch.ost.i.dsl.di_ex6;

import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

interface NotificationService {
    void sendNotification(String to, String message);
}

@Service("emailSender")  // Give it a name
class EmailSender implements NotificationService {
    @Override
    public void sendNotification(String to, String message) {
        System.out.println("📧 Email to " + to + ": " + message);
    }
}

@Service("smsSender")  // Give it a name
class SmsSender implements NotificationService {
    @Override
    public void sendNotification(String to, String message) {
        System.out.println("📱 SMS to " + to + ": " + message);
    }
}

@RestController
@RequestMapping("/api/v6")
class UserController {

    private final NotificationService notificationService;

    // @Autowired is optional on constructor since Spring 4.3
    public UserController(@Qualifier("smsSender") NotificationService notificationService) {
        this.notificationService = notificationService;
    }

    @PostMapping("/register")
    public String registerUser(@RequestParam String username) {
        System.out.println("Registering user: " + username);
        notificationService.sendNotification(username + "@example.com", "Welcome!");
        return "User registered: " + username;
    }
}
