import React, { useState } from 'react';
import { TextField, Button, Container, Typography, Box, IconButton, InputAdornment } from '@mui/material';
import { Visibility, VisibilityOff } from '@mui/icons-material';

function SentinelForm({ onSubmit }) {
    const [name, setName] = useState('');
    const [secretName, setSecretName] = useState('');
    const [password, setPassword] = useState('');
    const [secretType, setSecretType] = useState('');
    const [serviceAccount, setServiceAccount] = useState('');
    const [role, setRole] = useState('');
    const [roleBinding, setRoleBinding] = useState('');
    const [showPassword, setShowPassword] = useState(false); // State for password visibility

    const handleSubmit = async (e) => {
        e.preventDefault();
        const customResource = {
            apiVersion: 'secops.kavinduxo.com/v1alpha1',
            kind: 'Sentinel',
            metadata: {
                name: name.trim(),
                labels: {
                    usertype: 'ServiceAccount'
                }
            },
            spec: {
                secretName: secretName.trim(),
                data: {
                    password: password.trim()
                },
                secretType: secretType.trim(),
                serviceAccount: serviceAccount.trim(),
                role: role.trim(),
                roleBinding: roleBinding.trim()
            }
        };

        try {
            const response = await fetch('/api/v1/namespaces/default/sentinels', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(customResource),
            });

            if (!response.ok) {
                throw new Error('Failed to deploy Sentinel');
            }

            console.log('Sentinel deployed successfully!');
        } catch (error) {
            console.error('Error deploying sentinel:', error);
        }

    };

    return (
        <Container maxWidth="sm">
            <Box sx={{ marginTop: 4, textAlign: 'center' }}>
                <Typography variant="h4">Create Sentinel</Typography>
                <form onSubmit={handleSubmit}>
                    <TextField
                        label="Name"
                        variant="outlined"
                        fullWidth
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        sx={{ marginBottom: 2 }}
                    />
                    <TextField
                        label="Secret Name"
                        variant="outlined"
                        fullWidth
                        value={secretName}
                        onChange={(e) => setSecretName(e.target.value)}
                        sx={{ marginBottom: 2 }}
                    />
                    <TextField
                        label="Password"
                        variant="outlined"
                        fullWidth
                        type={showPassword ? 'text' : 'password'} // Toggle password visibility
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        InputProps={{
                            endAdornment: (
                                <InputAdornment position="end">
                                    <IconButton
                                        onClick={() => setShowPassword(!showPassword)} // Toggle show/hide password
                                        edge="end"
                                    >
                                        {showPassword ? <Visibility /> : <VisibilityOff />}
                                    </IconButton>
                                </InputAdornment>
                            ),
                        }}
                        sx={{ marginBottom: 2 }}
                    />
                    <TextField
                        label="Secret Type"
                        variant="outlined"
                        fullWidth
                        value={secretType}
                        onChange={(e) => setSecretType(e.target.value)}
                        sx={{ marginBottom: 2 }}
                    />
                    <TextField
                        label="Service Account"
                        variant="outlined"
                        fullWidth
                        value={serviceAccount}
                        onChange={(e) => setServiceAccount(e.target.value)}
                        sx={{ marginBottom: 2 }}
                    />
                    <TextField
                        label="Role"
                        variant="outlined"
                        fullWidth
                        value={role}
                        onChange={(e) => setRole(e.target.value)}
                        sx={{ marginBottom: 2 }}
                    />
                    <TextField
                        label="Role Binding"
                        variant="outlined"
                        fullWidth
                        value={roleBinding}
                        onChange={(e) => setRoleBinding(e.target.value)}
                        sx={{ marginBottom: 2 }}
                    />
                    <Button type="submit" variant="contained" color="primary">
                        Create
                    </Button>
                </form>
            </Box>
        </Container>
    );
}

export default SentinelForm;
