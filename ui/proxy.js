// proxy.js
const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');

const app = express();
const accessToken = 'eyJhbGciOiJSUzI1NiIsImtpZCI6InMwNXgyd2c0enJtcFFSd083VnNISEEzMl9RZVFhUWIzb1hEbjdSeTBod3MifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiXSwiZXhwIjoxNzExOTcxMjU1LCJpYXQiOjE3MTE5Njc2NTUsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImRhc2hib2FyZCIsInVpZCI6ImFhOTY1MzgzLWE5OTgtNDAzZC05NjVmLWM0NTJmNDY5NmIxNiJ9fSwibmJmIjoxNzExOTY3NjU1LCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDpkYXNoYm9hcmQifQ.UAhhhdL1PqrcnzslzdzgV4O8jGJzhJHMx1fAhiZ7Dwb809E4z9FEknPeCC9i9RFAZR6NKd5SAimo77hz4getJYH-T8I2Ld4KdHv1RtfIIODH3KpBemhsbjfAN0ysL2sddRgFTEPfWpc6CAni1jK5hyD6EbHjEc1vHlOSPxnUPYAYynZqRg47cWkurucsWG2lpsfao4gODDxgRNjnaMFXjqPCE-BAAwVvKhA4cNlv8usds2QrbvAjVjjlNu8AR_iIjzGCJGFOu1kCaOeII_QuMq95vUGdjwor5Z7CKQit4wg4FE0c1LY87vQYfWwDqcGk-OIMa6wdFuEt3UoIf66few';

const kubernetesApiProxy = createProxyMiddleware({
  target: 'https://127.0.0.1:35375', // Replace with your Kubernetes API server URL
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
