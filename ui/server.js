// server.js (Node.js + Express)
// not being used
const express = require('express');
const axios = require('axios');

const app = express();
const PORT = process.env.PORT || 5000;

// Define your Kubernetes API server URL
const KUBERNETES_API_URL = '';
const accessToken = '';

app.use(express.json());

app.post('/create-secret', async (req, res) => {
  try {
    const response = await axios.post(
      `${KUBERNETES_API_URL}/api/v1/namespaces/my-namespace/secrets`,
      req.body,
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${accessToken}`
        }
      }
    );

    res.json(response.data);
  } catch (error) {
    res.status(500).json({ error: 'Internal Server Error' });
  }
});

app.listen(PORT, () => {
  console.log(`Server is running on port ${PORT}`);
});
