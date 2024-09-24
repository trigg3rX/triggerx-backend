// Import necessary libraries
const TronWeb = require('tronweb');

const axios = require('axios');
 // Import Axios for API requests
 // Import Axios for API requests

// TRON network configuration
const TRON_NODE_URL = 'https://api.shasta.trongrid.io'; // TRON node URL
const PRIVATE_KEY = '7ef3fe1f4c142e76f823f851efbb1627f2d64f5b27af5fc99e9a1cd1d748c5bb'; // Replace with your actual private key

// Initialize TronWeb
const tronWeb = new TronWeb({
    fullHost: TRON_NODE_URL,
    privateKey: PRIVATE_KEY,
});

// Your deployed contract address
const CONTRACT_ADDRESS = 'TUKNarxTVvXdtgvLnNep6aahWuo5jMkm5h'; // Replace with your actual contract address

// Function to fetch the current USDC price from CoinGecko
async function fetchUSDCPrice() {
    try {
        const response = await axios.get('https://api.coingecko.com/api/v3/simple/price?ids=usd-coin&vs_currencies=usd');
        return response.data['usd-coin'].usd; // Return the USDC price in USD
    } catch (error) {
        console.error('Error fetching USDC price:', error);
        throw error; // Rethrow the error for handling later
    }
}

// Function to execute the task with the current USDC price
async function executeTaskWithCurrentPrice(taskId) {
    try {
        const usdcPrice = await fetchUSDCPrice(); // Fetch the current USDC price
        console.log('Current USDC price:', usdcPrice); // Log the price

        // Get contract instance
        const contract = await tronWeb.contract().at(CONTRACT_ADDRESS);

        // Encode the price (multiply by 1e18 for 18 decimal places and convert to hex)
        const encodedPrice = tronWeb.toBigNumber(usdcPrice * 1e18).toString(16).padStart(64, '0');

        // Execute the task
        const result = await contract.executeTask(taskId, '0x' + encodedPrice).send();

        console.log('Transaction successful:', result); // Log the result

    } catch (error) {
        console.error('Error executing task:', error); // Log any errors that occur
    }
}

// Main function to run the script
async function main() {
    const taskId = 1; // Replace with your actual task ID
    await executeTaskWithCurrentPrice(taskId); // Execute the task
}

// Run the main function
main().then(() => process.exit(0)).catch(error => {
    console.error(error); // Log any errors in the main function
    process.exit(1); // Exit the process with a failure code
});
