<script>
    const API_URL = "/api";

    let newProduct = $state({ name: "", description: "", price: null, stock: null });
    let productsPromise = $state(fetchProducts());

    // Fetch products
    async function fetchProducts() {
        const response = await fetch(`${API_URL}/products`);
        if (!response.ok) throw new Error("Failed to fetch products");
        return await response.json();
    }

    // Add product
    async function addProduct(event) {
        event.preventDefault();
        try {
            const response = await fetch(`${API_URL}/products`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(newProduct),
            });

            if (!response.ok) throw new Error("Failed to add product");

            // Reset form
            newProduct.name = "";
            newProduct.description = "";
            newProduct.price = null;
            newProduct.stock = null;

            // Refresh list
            productsPromise = fetchProducts();
        } catch (err) {
            alert("Error adding product: " + err.message);
        }
    }

    // Add unit
    async function buy(productId) {
        try {
            const response = await fetch(`${API_URL}/products/${productId}/buy`, {
                method: "POST",
            });

            if (!response.ok) throw new Error("Failed to buy product");

            // Refresh list
            productsPromise = fetchProducts();
        } catch (err) {
            alert("Error buying product: " + err.message);
        }
    }
</script>

<svelte:head>
    <title>Shop</title>
    <link
        rel="icon"
        href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 10'><circle cx='5' cy='5' r='4' fill='%23007bff'/></svg>"
    />
</svelte:head>

<h1>Shop - PrevelteKit Demo</h1>

<h2>Add New Product</h2>
<form onsubmit={addProduct}>
    <input
        type="text"
        bind:value={newProduct.name}
        placeholder="Product Name"
        required
    />
    <textarea
        bind:value={newProduct.description}
        placeholder="Description"
        required
    ></textarea>
    <input
        type="number"
        bind:value={newProduct.price}
        placeholder="Price (CHF)"
        min="0"
        step="0.01"
        required
    />
    <input
        type="number"
        bind:value={newProduct.stock}
        placeholder="Stock"
        min="0"
        step="1"
        required
    />
    <input type="submit" value="Add Product" />
</form>

<h2>Product List</h2>
{#await productsPromise}
    <div class="loading">Loading products...</div>
{:then products}
    {#if products.length > 0}
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Name</th>
                    <th>Description</th>
                    <th>Price</th>
                    <th>Stock</th>
                    <th></th>
                </tr>
            </thead>
            <tbody>
                {#each products as product (product.id)}
                    <tr>
                        <td>{product.id}</td>
                        <td>{product.name}</td>
                        <td>{product.description}</td>
                        <td>CHF {product.price.toFixed(2)}</td>
                        <td>{product.stock}</td>
                        <td>
                            <button disabled={product.stock === 0} onclick={() => buy(product.id)}>
                                {product.stock === 0 ? "Sold out" : "Buy"}
                            </button>
                        </td>
                    </tr>
                {/each}
            </tbody>
        </table>
    {:else}
        <p>No products found. Add one above!</p>
    {/if}
{:catch error}
    <div style="color: red;">Error: {error.message}</div>
{/await}

<style>
    :global(body) {
        font-family: Arial, sans-serif;
        max-width: 800px;
        margin: 50px auto;
        padding: 20px;
    }
    h1 {
        color: #333;
    }
    table {
        width: 100%;
        border-collapse: collapse;
        margin: 20px 0;
    }
    th,
    td {
        padding: 12px;
        text-align: left;
        border-bottom: 1px solid #ddd;
    }
    th {
        background-color: #4caf50;
        color: white;
    }
    form {
        background: #f4f4f4;
        padding: 20px;
        border-radius: 5px;
        margin: 20px 0;
    }
    input[type="text"],
    textarea {
        width: calc(100% - 20px);
        padding: 10px;
        margin: 5px 0;
        border: 1px solid #ddd;
        border-radius: 3px;
    }
    textarea {
        resize: vertical;
        min-height: 60px;
    }
    input[type="submit"] {
        background-color: #4caf50;
        color: white;
        padding: 10px 20px;
        border: none;
        border-radius: 3px;
        cursor: pointer;
        margin: 5px 0;
    }
    input[type="submit"]:hover {
        background-color: #45a049;
    }
    button {
        background-color: #4caf50;
        color: white;
        padding: 8px 16px;
        border: none;
        border-radius: 3px;
        cursor: pointer;
    }
    button:disabled {
        background-color: #aaa !important;
        cursor: not-allowed;
    }
    .loading {
        text-align: center;
        color: #666;
    }
</style>
