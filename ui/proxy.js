// proxy.js
const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
require('dotenv').config();

const app = express();
const url = process.env.URL;
const accessToken = process.env.TOKEN;

const kubernetesApiProxy = createProxyMiddleware({
  target: url, // Replace with your Kubernetes API server URL
  changeOrigin: true,
  secure: false, // Disable TLS certificate verification (not recommended for production)
  headers: {
    'Authorization': `Bearer ${accessToken}`,
  },
});

app.use('/apis', kubernetesApiProxy);

const PORT = process.env.PORT || 5000;
app.listen(PORT, () => {
  console.log(`Proxy server listening on port ${PORT}`);
});
