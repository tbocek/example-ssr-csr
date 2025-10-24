package ch.ost.i.dsl.ssr;

import jakarta.persistence.*;

@Entity
@Table(name = "games")
public class Game {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    private String title;
    private String description;
    private Integer stars = 0;
    
    public Game() {}
    
    public Game(String title, String description) {
        this.title = title;
        this.description = description;
        this.stars = 0;
    }
    
    public Long getId() {
        return id;
    }
    
    public void setId(Long id) {
        this.id = id;
    }
    
    public String getTitle() {
        return title;
    }
    
    public void setTitle(String title) {
        this.title = title;
    }
    
    public String getDescription() {
        return description;
    }
    
    public void setDescription(String description) {
        this.description = description;
    }
    
    public Integer getStars() {
        return stars;
    }
    
    public void setStars(Integer stars) {
        this.stars = stars;
    }
    
    public void addStar() {
        this.stars++;
    }
}