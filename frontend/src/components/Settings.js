import { useState, useEffect, useCallback } from 'react';
import {
    Box,
    Typography,
    Card,
    CardContent,
    CircularProgress,
    Grid,
} from '@mui/material';
import axios from 'axios';

function Settings({ token }) {
    const [config, setConfig] = useState(null);
    const [loading, setLoading] = useState(true);

    const fetchConfig = useCallback(async () => {
        try {
            const response = await axios.get('/api/config', {
                headers: { Authorization: `Bearer ${token}` },
            });
            setConfig(response.data);
        } catch (error) {
            console.error('Failed to fetch config:', error);
        } finally {
            setLoading(false);
        }
    }, [token]);

    useEffect(() => {
        fetchConfig();
    }, [fetchConfig]);

    if (loading) {
        return (
            <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
                <CircularProgress />
            </Box>
        );
    }

    return (
        <Box>
            <Typography variant="h4" gutterBottom>
                Settings
            </Typography>
            <Grid container spacing={3}>
                <Grid item xs={12} md={6}>
                    <Card>
                        <CardContent>
                            <Typography variant="h6" gutterBottom>
                                Server Configuration
                            </Typography>
                            {config && (
                                <Box>
                                    <Typography variant="body2" color="text.secondary" paragraph>
                                        Server Address: {config.server_addr}
                                    </Typography>
                                    <Typography variant="body2" color="text.secondary" paragraph>
                                        SMTP Host: {config.smtp_host || 'Not configured'}
                                    </Typography>
                                    <Typography variant="body2" color="text.secondary" paragraph>
                                        SMTP Port: {config.smtp_port || 'Not configured'}
                                    </Typography>
                                    <Typography variant="body2" color="text.secondary" paragraph>
                                        Email From: {config.email_from || 'Not configured'}
                                    </Typography>
                                    <Typography variant="body2" color="text.secondary" paragraph>
                                        Alert Email: {config.alert_email || 'Not configured'}
                                    </Typography>
                                </Box>
                            )}
                        </CardContent>
                    </Card>
                </Grid>
                <Grid item xs={12} md={6}>
                    <Card>
                        <CardContent>
                            <Typography variant="h6" gutterBottom>
                                Information
                            </Typography>
                            <Typography variant="body2" color="text.secondary" paragraph>
                                Version: 1.0.0
                            </Typography>
                            <Typography variant="body2" color="text.secondary" paragraph>
                                Edit server-config.json file to update server settings.
                            </Typography>
                            <Typography variant="body2" color="text.secondary" paragraph>
                                Restart the server after making configuration changes.
                            </Typography>
                        </CardContent>
                    </Card>
                </Grid>
            </Grid>
        </Box>
    );
}

export default Settings;
