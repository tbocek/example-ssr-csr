package ch.ost.i.dsl.ssr;

import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;

@Controller
public class ProductController {
    
    private final ProductRepository productRepository;
        
    public ProductController(ProductRepository productRepository) {
        this.productRepository = productRepository;
    }
        
    @GetMapping({"/", "/products"})
    public String listProducts(Model model) {
        model.addAttribute("products", productRepository.findAll());
        model.addAttribute("newProduct", new Product());
        return "products";
    }
    
    @PostMapping("/products")
    public String createProduct(@ModelAttribute Product product) {
        productRepository.save(product);
        return "redirect:/products";
    }
    
    @PostMapping("/products/{id}/buy")
    public String buy(@PathVariable Long id) {
        productRepository.findById(id).ifPresent(product -> {
            product.buy();
            productRepository.save(product);
        });
        return "redirect:/products";
    }
}
