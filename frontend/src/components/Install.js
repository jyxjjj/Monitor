import { useState, useEffect, useCallback } from 'react';
import {
    Container,
    Paper,
    Button,
    Typography,
    Box,
    Alert,
    CircularProgress,
    Stepper,
    Step,
    StepLabel,
} from '@mui/material';
import axios from 'axios';

function Install({ onInstalled }) {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState(false);
    const [checking, setChecking] = useState(true);
    const [installed, setInstalled] = useState(false);

    const checkInstallation = useCallback(async () => {
        try {
            const response = await axios.get('/api/install/check');
            setInstalled(response.data.installed);
            if (response.data.installed && onInstalled) {
                onInstalled();
            }
        } catch (err) {
            console.error('Failed to check installation:', err);
        } finally {
            setChecking(false);
        }

    }, [onInstalled]);

    useEffect(() => {
        checkInstallation();
    }, [checkInstallation]);

    const handleInstall = async () => {
        setError('');
        setLoading(true);

        try {
            const response = await axios.post('/api/install/setup');
            if (response.data.success) {
                setSuccess(true);
                setTimeout(() => {
                    if (onInstalled) {
                        onInstalled();
                    }
                }, 2000);
            }
        } catch (err) {
            setError(err.response?.data || 'Installation failed');
        } finally {
            setLoading(false);
        }
    };

    if (checking) {
        return (
            <Container maxWidth="sm">
                <Box
                    sx={{
                        marginTop: 8,
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                    }}
                >
                    <CircularProgress />
                    <Typography variant="body1" sx={{ mt: 2 }}>
                        Checking installation status...
                    </Typography>
                </Box>
            </Container>
        );
    }

    if (installed) {
        return null;
    }

    return (
        <Container maxWidth="md">
            <Box
                sx={{
                    marginTop: 8,
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                }}
            >
                <Paper elevation={3} sx={{ padding: 4, width: '100%' }}>
                    <Typography component="h1" variant="h4" align="center" gutterBottom>
                        Monitor Installation
                    </Typography>
                    <Typography variant="body1" align="center" color="textSecondary" gutterBottom sx={{ mb: 4 }}>
                        Welcome to Monitor! Click the button below to initialize the database.
                    </Typography>

                    <Stepper activeStep={success ? 1 : 0} sx={{ mb: 4 }}>
                        <Step>
                            <StepLabel>Initialize Database</StepLabel>
                        </Step>
                        <Step>
                            <StepLabel>Ready to Use</StepLabel>
                        </Step>
                    </Stepper>

                    {error && (
                        <Alert severity="error" sx={{ mb: 3 }}>
                            {error}
                        </Alert>
                    )}

                    {success && (
                        <Alert severity="success" sx={{ mb: 3 }}>
                            Database installed successfully! Redirecting...
                        </Alert>
                    )}

                    <Box sx={{ mb: 3 }}>
                        <Typography variant="h6" gutterBottom>
                            What will be installed:
                        </Typography>
                        <Typography variant="body2" color="textSecondary" paragraph>
                            • Agents table - Store monitored server information
                        </Typography>
                        <Typography variant="body2" color="textSecondary" paragraph>
                            • Metrics table - Store historical performance data
                        </Typography>
                        <Typography variant="body2" color="textSecondary" paragraph>
                            • Alert Rules table - Store alert configurations
                        </Typography>
                        <Typography variant="body2" color="textSecondary" paragraph>
                            • Alerts table - Store alert history
                        </Typography>
                    </Box>

                    <Button
                        fullWidth
                        variant="contained"
                        onClick={handleInstall}
                        disabled={loading || success}
                        size="large"
                    >
                        {loading ? (
                            <>
                                <CircularProgress size={24} sx={{ mr: 1 }} />
                                Installing...
                            </>
                        ) : success ? (
                            'Installation Complete'
                        ) : (
                            'Install Database'
                        )}
                    </Button>

                    <Typography variant="caption" color="textSecondary" align="center" display="block" sx={{ mt: 2 }}>
                        Note: Make sure your database server is running and accessible
                    </Typography>
                </Paper>
            </Box>
        </Container>
    );
}

export default Install;
