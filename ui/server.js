// server.js (Node.js + Express)

const express = require('express');
const axios = require('axios');

const app = express();
const PORT = process.env.PORT || 5000;

// Define your Kubernetes API server URL
const KUBERNETES_API_URL = 'https://127.0.0.1:36245';
const accessToken = 'eyJhbGciOiJSUzI1NiIsImtpZCI6Im9vWENNR0dNRlprQmQyUEYtaUxudFZhQmhsTHFuQWJBZkhkZm9BRWYxZWMifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlZmF1bHQtdG9rZW4iLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVmYXVsdCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImZkMmI2ZGYyLWM4OWEtNDI1Mi1hYzAyLWQ3NWQ0YzBjNDgyZiIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlZmF1bHQifQ.BNcf-TJHfyfxBEmsOyMzwy8jORgIklIzxDopNmcQG3tEsI_U_XyqoQhmx8zKgr8mOb4KZKSmJT82JqWU0Rwaja81HIutWzFRdgXXyYUo8hCLj5lNWV4pKJnqsOcwFf93eyB9AjPxDslvwtOhHex0tV4DGMMsL7WQRKcm9dgA2FKrtmMaPLhk9_B0QY8aJUbKAGbxei5cdAdBUYzpsHvZEx0cY7PkPwCcQDs1gYwRYdhWNwC1S8s6667hzzmxisjzoHTApBzncRyM2CBqVl6jdo-56yvsQRAXlv0uMHT6LE8lGQNPu7V1EytdgCJe_fXFpnRLPdc_xnVJRqjuQpp_Aw';

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
