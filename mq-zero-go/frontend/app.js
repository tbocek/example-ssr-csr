const { createApp } = Vue;

const API_URL = '/api';

createApp({
    data() {
        return {
            products: [],
            newProduct: {
                name: '',
                description: '',
                price: null,
                stock: null
            },
            loading: true,
            error: null
        }
    },
    mounted() {
        this.fetchProducts();
    },
    methods: {
        async fetchProducts() {
            try {
                const response = await fetch(`${API_URL}/products`);
                if (!response.ok) throw new Error('Failed to fetch products');
                this.products = await response.json();
                this.loading = false;
            } catch (err) {
                this.error = err.message;
                this.loading = false;
            }
        },
        async addProduct() {
            try {
                const response = await fetch(`${API_URL}/products`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(this.newProduct)
                });
                
                if (!response.ok) throw new Error('Failed to add product');
                
                // Reset form
                this.newProduct.name = '';
                this.newProduct.description = '';
                this.newProduct.price = null;
                this.newProduct.stock = null;
                
                // Refresh list
                await this.fetchProducts();
            } catch (err) {
                alert('Error adding product: ' + err.message);
            }
        },
        async buy(productId) {
            try {
                const response = await fetch(`${API_URL}/products/${productId}/buy`, {
                    method: 'POST'
                });
                
                if (!response.ok) throw new Error('Failed to buy product');
                
                // Update the product's stock locally
                const product = this.products.find(g => g.id === productId);
                if (product) {
                    product.stock--;
                }
            } catch (err) {
                alert('Error buying product: ' + err.message);
            }
        }
    }
}).mount('#app');