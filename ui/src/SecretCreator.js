// SecretCreator.js (React component)

import React from 'react';
import axios from 'axios';

const SecretCreator = () => {
  const createSecret = async () => {
    try {
      const secretData = {
        // Define your secret data
      };

      const response = await axios.post(
        '/create-secret', // Send the request to your backend endpoint
        secretData
      );

      console.log('Secret created:', response.data);
    } catch (error) {
      console.error('Error creating secret:', error);
    }
  };

  return (
    <div>
      <button onClick={createSecret}>Create Secret</button>
    </div>
  );
};

export default SecretCreator;
