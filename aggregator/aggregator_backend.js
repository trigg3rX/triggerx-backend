const express = require('express');
const aggregatorApp = express();
aggregatorApp.use(express.json());

aggregatorApp.post('/receive-result', (req, res) => {
    const { jobId, result } = req.body;
    console.log(`Aggregator received result for job ${jobId}:`, result);

    // Here you can store the result or trigger further actions

    res.status(200).send(`Result for job ${jobId} received and processed.`);
});

const aggregatorPort = 3002;
aggregatorApp.listen(aggregatorPort, () => {
    console.log(`Aggregator service listening at http://localhost:${aggregatorPort}`);
});
