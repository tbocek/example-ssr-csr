package ch.ost.i.dsl.tx;

import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;

@Controller
public class ProductController {
    
    private final ProductRepository productRepository;
    private final ProductService productService;  // Use service layer for transactional operations
        
    public ProductController(ProductRepository productRepository, ProductService productService) {
        this.productRepository = productRepository;
        this.productService = productService;
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
        try {
            productService.buyWithStatistics(id);
        } catch (IllegalArgumentException e) {
            // Handle product not found
            return "redirect:/products?error=not-found";
        } catch (IllegalStateException e) {
            return "redirect:/products?error=sold-out";
        }
        return "redirect:/products";
    }
    
    @PostMapping("/products/{fromId}/transfer/{toId}")
    public String transferStock(@PathVariable Long fromId, 
                               @PathVariable Long toId,
                               @ModelAttribute("stock") int stock) {
        try {
            productService.transferStock(fromId, toId, stock);
        } catch (IllegalStateException e) {
            return "redirect:/products?error=insufficient-stock";
        }
        return "redirect:/products";
    }
}