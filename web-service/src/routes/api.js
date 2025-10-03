const express = require('express');
const axios = require('axios');

const router = express.Router();

router.post('/trigger', async (req, res) => {
    try {
        const response = await axios.get('http://api-service-1:8080/api/data');
        res.json({ result: response.data });
    } catch (error) {
        res.status(500).json({ error: 'An error occurred while processing your request.' });
    }
});

module.exports = router;