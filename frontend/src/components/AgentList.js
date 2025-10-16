import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import {
  Grid,
  Card,
  CardContent,
  Typography,
  Chip,
  Box,
  CircularProgress,
} from '@mui/material';
import {
  Computer as ComputerIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
} from '@mui/icons-material';
import axios from 'axios';

function AgentList({ token }) {
  const [agents, setAgents] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAgents();
    const interval = setInterval(fetchAgents, 5000);
    return () => clearInterval(interval);
  }, [token]);

  const fetchAgents = async () => {
    try {
      const response = await axios.get('/api/agents', {
        headers: { Authorization: `Bearer ${token}` },
      });
      setAgents(response.data || []);
    } catch (error) {
      console.error('Failed to fetch agents:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  if (agents.length === 0) {
    return (
      <Box textAlign="center" mt={4}>
        <ComputerIcon sx={{ fontSize: 60, color: 'text.secondary', mb: 2 }} />
        <Typography variant="h6" color="text.secondary">
          No agents connected
        </Typography>
        <Typography variant="body2" color="text.secondary" mt={1}>
          Start an agent to see it here
        </Typography>
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Agents
      </Typography>
      <Grid container spacing={3}>
        {agents.map((agent) => (
          <Grid item xs={12} sm={6} md={4} key={agent.id}>
            <Card
              component={Link}
              to={`/agents/${agent.id}`}
              sx={{
                textDecoration: 'none',
                '&:hover': {
                  boxShadow: 6,
                },
              }}
            >
              <CardContent>
                <Box display="flex" alignItems="center" mb={2}>
                  <ComputerIcon sx={{ mr: 1, color: 'primary.main' }} />
                  <Typography variant="h6" component="div">
                    {agent.name}
                  </Typography>
                </Box>
                <Box mb={1}>
                  <Chip
                    label={agent.status}
                    color={agent.status === 'online' ? 'success' : 'error'}
                    size="small"
                    icon={agent.status === 'online' ? <CheckCircleIcon /> : <ErrorIcon />}
                  />
                  <Chip
                    label={agent.platform}
                    size="small"
                    sx={{ ml: 1 }}
                  />
                </Box>
                <Typography variant="body2" color="text.secondary">
                  Host: {agent.host}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Last seen: {new Date(agent.last_seen).toLocaleString()}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Box>
  );
}

export default AgentList;
