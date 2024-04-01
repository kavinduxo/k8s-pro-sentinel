import React, { useState } from 'react';
import { TextField, Button, Container, Typography, Box, IconButton, InputAdornment, Select, MenuItem, Snackbar } from '@mui/material';
import { Visibility, VisibilityOff } from '@mui/icons-material';
import MuiAlert from '@mui/material/Alert';
import Logo from './resources/pro-logo.png'; // Import your logo image

function Alert(props) {
    return <MuiAlert elevation={6} variant="filled" {...props} />;
}

function SentinelForm({ onSubmit }) {
    const [name, setName] = useState('');
    const [secretName, setSecretName] = useState('');
    const [password, setPassword] = useState('');
    const [secretType, setSecretType] = useState('BaseSecret');
    const [serviceAccount, setServiceAccount] = useState('');
    const [role, setRole] = useState('');
    const [roleBinding, setRoleBinding] = useState('');
    const [showPassword, setShowPassword] = useState(false); // State for password visibility
    const [snackbarOpen, setSnackbarOpen] = useState(false); // State for controlling Snackbar open/close
    const [snackbarMessage, setSnackbarMessage] = useState(''); // State for Snackbar message
    const [snackbarSeverity, setSnackbarSeverity] = useState('success'); // State for Snackbar severity (success/error)

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
            const response = await fetch('/apis/secops.kavinduxo.com/v1alpha1/namespaces/default/sentinels', {
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
            clearFormFields();
            setSnackbarSeverity('success');
            setSnackbarMessage('Sentinel deployed successfully!');
            setSnackbarOpen(true);
        } catch (error) {
            console.error('Error deploying sentinel:', error);
            setSnackbarSeverity('error');
            setSnackbarMessage('Failed to deploy Sentinel');
            setSnackbarOpen(true);
        }
    };

    const clearFormFields = () => {
        setName('');
        setSecretName('');
        setPassword('');
        setSecretType('BaseSecret');
        setServiceAccount('');
        setRole('');
        setRoleBinding('');
    };

    const handleSnackbarClose = () => {
        setSnackbarOpen(false);
    };

    return (
        <Container maxWidth="sm">
            <Box sx={{ marginTop: 3, marginBottom: 3, display: 'flex', alignItems: 'center', textAlign: 'center' }}>
                <Box sx={{ marginRight: 4, paddingLeft: 15 }}>
                    <img src={Logo} alt="Logo" style={{ width: 75 }} />
                </Box>
                <Box>
                    <Typography variant="h5" component="h1" sx={{ fontWeight: 'bold', marginBottom: 1, fontFamily: 'Arial, sans-serif' }}>Create Sentinel</Typography>
                </Box>
            </Box>
            <Box sx={{ textAlign: 'center' }}>
                <form onSubmit={handleSubmit}>
                    <Box sx={{ borderBottom: '2px solid #ccc', marginBottom: 1 }}>
                        <TextField
                            label="Name"
                            variant="outlined"
                            fullWidth
                            size="small"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            sx={{ marginBottom: 1 }}
                        />
                    </Box>
                    <Box sx={{ borderBottom: '2px solid #ccc', marginBottom: 1 }}>
                        <Typography variant="h6" sx={{ marginBottom: 1 }}>Credentials</Typography>
                        <TextField
                            label="Secret Name"
                            variant="outlined"
                            fullWidth
                            size="small"
                            value={secretName}
                            onChange={(e) => setSecretName(e.target.value)}
                            sx={{ marginBottom: 1 }}
                        />
                        <TextField
                            label="Password"
                            variant="outlined"
                            fullWidth
                            size="small"
                            type={showPassword ? 'text' : 'password'}
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            InputProps={{
                                endAdornment: (
                                    <InputAdornment position="end">
                                        <IconButton
                                            onClick={() => setShowPassword(!showPassword)}
                                            edge="end"
                                        >
                                            {showPassword ? <Visibility /> : <VisibilityOff />}
                                        </IconButton>
                                    </InputAdornment>
                                ),
                            }}
                            sx={{ marginBottom: 1 }}
                        />
                        <Select
                            value={secretType}
                            onChange={(e) => setSecretType(e.target.value)}
                            variant="outlined"
                            fullWidth
                            size="small"
                            sx={{
                                marginBottom: 1,
                                '& .MuiSelect-root': {
                                    borderRadius: 0,
                                    borderBottom: '2px solid #ccc',
                                },
                                '& .MuiSelect-icon': {
                                    color: '#000',
                                },
                            }}
                        >
                            <MenuItem value="BaseSecret">Base Secret</MenuItem>
                            <MenuItem value="RbacBaseSecret">Rbac-Base Secret</MenuItem>
                            <MenuItem value="SecuredSecret">Secured Secret</MenuItem>
                            <MenuItem value="RbacSecuredSecret">Rbac-Secured Secret</MenuItem>
                            <MenuItem value="KMSSecuredSecret">KMS Secured Secret</MenuItem>
                            <MenuItem value="RbacKMSSecuredSecret">Rbac-KMS Secured Secret</MenuItem>
                        </Select>
                    </Box>
                    {(secretType === 'RbacBaseSecret' || secretType === 'RbacSecuredSecret' || secretType === 'RbacKMSSecuredSecret') && (
                        <Box sx={{ borderBottom: '2px solid #ccc', marginBottom: 3 }}>
                            <Typography variant="h6" sx={{ marginBottom: 1 }}>Authorization</Typography>
                            <TextField
                                label="Service Account"
                                variant="outlined"
                                fullWidth
                                size="small"
                                value={serviceAccount}
                                onChange={(e) => setServiceAccount(e.target.value)}
                                sx={{ marginBottom: 1 }}
                            />
                            <TextField
                                label="Role"
                                variant="outlined"
                                fullWidth
                                size="small"
                                value={role}
                                onChange={(e) => setRole(e.target.value)}
                                sx={{ marginBottom: 1 }}
                            />
                            <TextField
                                label="Role Binding"
                                variant="outlined"
                                fullWidth
                                size="small"
                                value={roleBinding}
                                onChange={(e) => setRoleBinding(e.target.value)}
                                sx={{ marginBottom: 1 }}
                            />
                        </Box>
                    )}
                    <Box sx={{ display: 'flex', justifyContent: 'center', marginBottom: 2 }}>
                        <Button type="submit" variant="contained" color="primary" sx={{ marginRight: 1 }}>
                            Create
                        </Button>
                        <Button variant="contained" color="error" onClick={clearFormFields} sx={{ marginRight: 1 }}>
                            Clear
                        </Button>
                        <Button variant="outlined" color="primary" href="https://github.com/kavinduxo/k8s-pro-sentinel#k8s-pro-sentinel-kubernetes-security-operator" target="_blank">
                            Go to Documentation
                        </Button>
                    </Box>
                </form>
                <Snackbar open={snackbarOpen} autoHideDuration={6000} onClose={handleSnackbarClose}>
                    <div>
                        <Alert onClose={handleSnackbarClose} severity={snackbarSeverity}>
                            {snackbarMessage}
                        </Alert>
                    </div>
                </Snackbar>
            </Box>
        </Container>
    );
}

export default SentinelForm;
