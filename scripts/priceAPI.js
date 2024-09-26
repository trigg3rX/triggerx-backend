const express = require('express');

const app = express();
const port = 3003;

// Simulated price API endpoint
app.get('/get-price', (req, res) => {
    // Example price, this can be replaced with dynamic logic
    const price = {
        data: {
            value: 45670987665
        }
    };
    res.json(price);
});

// Start the price API server
app.listen(port, () => {
    console.log(`Price API listening at http://localhost:${port}`);
});
