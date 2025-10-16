import { useState, useEffect, useCallback } from 'react';
import {
    Box,
    Typography,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Paper,
    Button,
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    TextField,
    MenuItem,
    Switch,
    FormControlLabel,
    CircularProgress,
} from '@mui/material';
import axios from 'axios';

function AlertRules({ token }) {
    const [rules, setRules] = useState([]);
    const [loading, setLoading] = useState(true);
    const [openDialog, setOpenDialog] = useState(false);
    const [newRule, setNewRule] = useState({
        agent_id: '',
        metric_type: 'cpu',
        threshold: 80,
        operator: 'gt',
        duration: 60,
        enabled: true,
        description: '',
    });

    const fetchRules = useCallback(async () => {
        try {
            const response = await axios.get('/api/alert-rules', {
                headers: { Authorization: `Bearer ${token}` },
            });
            setRules(response.data || []);
        } catch (error) {
            console.error('Failed to fetch rules:', error);
        } finally {
            setLoading(false);
        }
    }, [token]);

    useEffect(() => {
        fetchRules();
    }, [fetchRules]);

    const handleAddRule = async () => {
        try {
            await axios.post('/api/alert-rules', newRule, {
                headers: { Authorization: `Bearer ${token}` },
            });
            setOpenDialog(false);
            fetchRules();
            setNewRule({
                agent_id: '',
                metric_type: 'cpu',
                threshold: 80,
                operator: 'gt',
                duration: 60,
                enabled: true,
                description: '',
            });
        } catch (error) {
            console.error('Failed to add rule:', error);
        }
    };

    if (loading) {
        return (
            <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
                <CircularProgress />
            </Box>
        );
    }

    return (
        <Box>
            <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
                <Typography variant="h4">
                    Alert Rules
                </Typography>
                <Button variant="contained" onClick={() => setOpenDialog(true)}>
                    Add Rule
                </Button>
            </Box>

            <TableContainer component={Paper}>
                <Table>
                    <TableHead>
                        <TableRow>
                            <TableCell>Agent</TableCell>
                            <TableCell>Metric</TableCell>
                            <TableCell>Condition</TableCell>
                            <TableCell>Duration</TableCell>
                            <TableCell>Description</TableCell>
                            <TableCell>Enabled</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {rules.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={6} align="center">
                                    <Typography color="text.secondary">No alert rules</Typography>
                                </TableCell>
                            </TableRow>
                        ) : (
                            rules.map((rule) => (
                                <TableRow key={rule.id}>
                                    <TableCell>{rule.agent_id || 'All'}</TableCell>
                                    <TableCell>{rule.metric_type}</TableCell>
                                    <TableCell>
                                        {rule.operator} {rule.threshold}
                                    </TableCell>
                                    <TableCell>{rule.duration}s</TableCell>
                                    <TableCell>{rule.description}</TableCell>
                                    <TableCell>{rule.enabled ? 'Yes' : 'No'}</TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </TableContainer>

            <Dialog open={openDialog} onClose={() => setOpenDialog(false)}>
                <DialogTitle>Add Alert Rule</DialogTitle>
                <DialogContent>
                    <TextField
                        margin="dense"
                        label="Agent ID (leave empty for all)"
                        type="text"
                        fullWidth
                        value={newRule.agent_id}
                        onChange={(e) => setNewRule({ ...newRule, agent_id: e.target.value })}
                    />
                    <TextField
                        margin="dense"
                        label="Metric Type"
                        select
                        fullWidth
                        value={newRule.metric_type}
                        onChange={(e) => setNewRule({ ...newRule, metric_type: e.target.value })}
                    >
                        <MenuItem value="cpu">CPU</MenuItem>
                        <MenuItem value="memory">Memory</MenuItem>
                        <MenuItem value="disk">Disk</MenuItem>
                        <MenuItem value="load">Load Average</MenuItem>
                    </TextField>
                    <TextField
                        margin="dense"
                        label="Operator"
                        select
                        fullWidth
                        value={newRule.operator}
                        onChange={(e) => setNewRule({ ...newRule, operator: e.target.value })}
                    >
                        <MenuItem value="gt">Greater Than</MenuItem>
                        <MenuItem value="lt">Less Than</MenuItem>
                        <MenuItem value="gte">Greater Than or Equal</MenuItem>
                        <MenuItem value="lte">Less Than or Equal</MenuItem>
                    </TextField>
                    <TextField
                        margin="dense"
                        label="Threshold"
                        type="number"
                        fullWidth
                        value={newRule.threshold}
                        onChange={(e) => setNewRule({ ...newRule, threshold: parseFloat(e.target.value) })}
                    />
                    <TextField
                        margin="dense"
                        label="Duration (seconds)"
                        type="number"
                        fullWidth
                        value={newRule.duration}
                        onChange={(e) => setNewRule({ ...newRule, duration: parseInt(e.target.value) })}
                    />
                    <TextField
                        margin="dense"
                        label="Description"
                        type="text"
                        fullWidth
                        multiline
                        rows={2}
                        value={newRule.description}
                        onChange={(e) => setNewRule({ ...newRule, description: e.target.value })}
                    />
                    <FormControlLabel
                        control={
                            <Switch
                                checked={newRule.enabled}
                                onChange={(e) => setNewRule({ ...newRule, enabled: e.target.checked })}
                            />
                        }
                        label="Enabled"
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
                    <Button onClick={handleAddRule} variant="contained">
                        Add
                    </Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
}

export default AlertRules;
