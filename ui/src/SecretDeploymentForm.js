// SecretDeploymentForm.js
import React, { useState } from 'react';

const SecretDeploymentForm = () => {
  const [secretData, setSecretData] = useState({});

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setSecretData({ ...secretData, [name]: value });
  };

  const deploySecret = async () => {
    try {
      const response = await fetch('/api/v1/namespaces/default/secrets', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          apiVersion: 'v1',
          kind: 'Secret',
          metadata: {
            name: 'ui-secret',
          },
          data: {
            username: "dXNlcm5hbWU=", // base64 encoded username
            password: "cGFzc3dvcmQ="  // base64 encoded password
          }
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to deploy secret');
      }

      console.log('Secret deployed successfully!');
    } catch (error) {
      console.error('Error deploying secret:', error);
    }
  };

  return (
    <div>
      <input type="text" name="username" placeholder="Username" onChange={handleInputChange} />
      <input type="password" name="password" placeholder="Password" onChange={handleInputChange} />
      <button onClick={deploySecret}>Deploy Secret</button>
    </div>
  );
};

export default SecretDeploymentForm;
