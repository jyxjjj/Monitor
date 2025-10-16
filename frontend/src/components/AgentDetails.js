import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import {
  Box,
  Typography,
  Grid,
  Card,
  CardContent,
  CircularProgress,
} from '@mui/material';
import { LineChart } from '@mui/x-charts/LineChart';
import axios from 'axios';

function AgentDetails({ token }) {
  const { agentId } = useParams();
  const [metrics, setMetrics] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(fetchMetrics, 5000);
    return () => clearInterval(interval);
  }, [agentId, token]);

  const fetchMetrics = async () => {
    try {
      const since = new Date(Date.now() - 3600000).toISOString(); // Last hour
      const response = await axios.get(`/api/metrics/${agentId}?since=${since}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = response.data || [];
      setMetrics(data.reverse());
    } catch (error) {
      console.error('Failed to fetch metrics:', error);
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

  const latest = metrics[metrics.length - 1];

  const timestamps = metrics.map((m) => new Date(m.timestamp));
  const cpuData = metrics.map((m) => m.cpu_percent);
  const memoryData = metrics.map((m) => (m.memory_used / m.memory_total) * 100);
  const diskData = metrics.map((m) => (m.disk_used / m.disk_total) * 100);
  const loadData = metrics.map((m) => m.load_avg_1);

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Agent: {agentId}
      </Typography>

      {latest && (
        <Grid container spacing={3} mb={3}>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Typography color="text.secondary" gutterBottom>
                  CPU Usage
                </Typography>
                <Typography variant="h4">
                  {latest.cpu_percent.toFixed(1)}%
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Typography color="text.secondary" gutterBottom>
                  Memory Usage
                </Typography>
                <Typography variant="h4">
                  {((latest.memory_used / latest.memory_total) * 100).toFixed(1)}%
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {formatBytes(latest.memory_used)} / {formatBytes(latest.memory_total)}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Typography color="text.secondary" gutterBottom>
                  Disk Usage
                </Typography>
                <Typography variant="h4">
                  {((latest.disk_used / latest.disk_total) * 100).toFixed(1)}%
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {formatBytes(latest.disk_used)} / {formatBytes(latest.disk_total)}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Typography color="text.secondary" gutterBottom>
                  Load Average (1m)
                </Typography>
                <Typography variant="h4">
                  {latest.load_avg_1.toFixed(2)}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  5m: {latest.load_avg_5.toFixed(2)} | 15m: {latest.load_avg_15.toFixed(2)}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}

      {metrics.length > 0 && (
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  CPU Usage (%)
                </Typography>
                <LineChart
                  xAxis={[{ data: timestamps, scaleType: 'time' }]}
                  series={[{ data: cpuData, label: 'CPU %' }]}
                  height={300}
                />
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={6}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Memory Usage (%)
                </Typography>
                <LineChart
                  xAxis={[{ data: timestamps, scaleType: 'time' }]}
                  series={[{ data: memoryData, label: 'Memory %' }]}
                  height={300}
                />
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={6}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Disk Usage (%)
                </Typography>
                <LineChart
                  xAxis={[{ data: timestamps, scaleType: 'time' }]}
                  series={[{ data: diskData, label: 'Disk %' }]}
                  height={300}
                />
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={6}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Load Average (1m)
                </Typography>
                <LineChart
                  xAxis={[{ data: timestamps, scaleType: 'time' }]}
                  series={[{ data: loadData, label: 'Load' }]}
                  height={300}
                />
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}
    </Box>
  );
}

export default AgentDetails;
