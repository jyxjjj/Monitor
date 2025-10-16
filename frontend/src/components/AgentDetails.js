import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import {
    Box,
    Typography,
    Card,
    CardContent,
    CircularProgress,
} from '@mui/material';
import { FormControl, InputLabel, Select, MenuItem } from '@mui/material';
import { LineChart } from '@mui/x-charts/LineChart';
import axios from 'axios';

/* eslint-disable no-extend-native */
// Override Date.prototype.toString so chart tooltips that rely on Date->string
// display local time 'YYYY-MM-DD HH:mm:ss'.
if (typeof Date !== 'undefined' && Date.prototype && Date.prototype.toString) {
    Date.prototype.toString = function () {
        const pad = (n) => String(n).padStart(2, '0');
        return `${this.getFullYear()}-${pad(this.getMonth() + 1)}-${pad(this.getDate())} ${pad(this.getHours())}:${pad(this.getMinutes())}:${pad(this.getSeconds())}`;
    };
}
/* eslint-enable no-extend-native */

function AgentDetails({ token }) {
    const { agentId } = useParams();
    const [metrics, setMetrics] = useState([]);
    const [loading, setLoading] = useState(true);
    const [range, setRange] = useState('5m');

    const handleRangeChange = (e) => {
        setRange(e.target.value);
    };

    const fetchMetrics = useCallback(async () => {
        try {
            // compute since based on selected range
            const now = Date.now();
            let sinceTs = now - 5 * 60 * 1000; // default 5m
            switch (range) {
                case '5m': sinceTs = now - 5 * 60 * 1000; break;
                case '30m': sinceTs = now - 30 * 60 * 1000; break;
                case '15m': sinceTs = now - 15 * 60 * 1000; break;
                case '1h': sinceTs = now - 60 * 60 * 1000; break;
                case '6h': sinceTs = now - 6 * 60 * 60 * 1000; break;
                case '24h': sinceTs = now - 24 * 60 * 60 * 1000; break;
                case '7d': sinceTs = now - 7 * 24 * 60 * 60 * 1000; break;
                default: sinceTs = now - 5 * 60 * 1000; break;
            }
            const pad = (n) => String(n).padStart(2, '0');
            const d = new Date(sinceTs);
            const since = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
            // Only send time-based since parameter (RFC3339). Server computes
            // aggregation buckets based on the time window; clients should not
            // request a specific points count anymore.
            const response = await axios.get(`/api/metrics/${agentId}?since=${since}`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            const data = response.data || [];
            // server returns ascending-ordered metrics (oldest -> newest)
            setMetrics(data);
        } catch (error) {
            console.error('Failed to fetch metrics:', error);
        } finally {
            setLoading(false);
        }
    }, [agentId, token, range]);

    useEffect(() => {
        fetchMetrics();
        const interval = setInterval(fetchMetrics, 5000);
        return () => clearInterval(interval);
    }, [fetchMetrics, range]);

    if (loading) {
        return (
            <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
                <CircularProgress />
            </Box>
        );
    }

    const latest = metrics[metrics.length - 1];

    const timestamps = metrics.map((m) => new Date(m.timestamp));

    // Determine expected reporting interval from data (median of deltas)
    const deltas = [];
    for (let i = 1; i < timestamps.length; i++) {
        deltas.push(timestamps[i] - timestamps[i - 1]);
    }
    const median = (arr) => {
        if (arr.length === 0) return 0;
        const s = [...arr].sort((a, b) => a - b);
        const mid = Math.floor(s.length / 2);
        return s.length % 2 === 0 ? (s[mid - 1] + s[mid]) / 2 : s[mid];
    };
    const expectedInterval = median(deltas) || 5000; // fallback to 5s
    const gapThreshold = Math.max(expectedInterval * 3, 60 * 1000); // 3x or at least 1 minute

    // Helper to create series with gaps (null) where timestamps have large gaps
    const buildSeriesWithGaps = (values) => {
        const out = [...values];
        for (let i = 1; i < values.length; i++) {
            if (timestamps[i] - timestamps[i - 1] > gapThreshold) {
                // mark the point after the gap as null so the line breaks and
                // we avoid creating an isolated hoverable empty point before the gap
                out[i] = null;
            }
        }
        return out;
    };

    const cpuRaw = metrics.map((m) => m.cpu_percent);
    const memoryRaw = metrics.map((m) => (m.memory_used / m.memory_total) * 100);
    const diskRaw = metrics.map((m) => (m.disk_used / m.disk_total) * 100);
    const loadRaw = metrics.map((m) => m.load_avg_1);

    const cpuData = buildSeriesWithGaps(cpuRaw);
    const memoryData = buildSeriesWithGaps(memoryRaw);
    const diskData = buildSeriesWithGaps(diskRaw);
    const loadData = buildSeriesWithGaps(loadRaw);

    // Compute Y axis max based on visible (non-null) values
    const computeMax = (arr) => {
        const vals = arr.filter((v) => v != null && !isNaN(v));
        if (vals.length === 0) return undefined;
        const mx = Math.max(...vals);
        // add small headroom
        return mx <= 0 ? 1 : Math.ceil(mx * 1.05);
    };
    const cpuMax = computeMax(cpuData);
    const memMax = computeMax(memoryData);
    const diskMax = computeMax(diskData);
    const loadMax = computeMax(loadData);

    // Determine x-axis tick formatter based on range
    const pad = (n) => String(n).padStart(2, '0');
    // Format time as 'YYYY-MM-DD HH:mm:ss' using local time
    const formatTick = (date) => {
        const d = new Date(date);
        return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
    };

    const formatBytes = (bytes) => {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    return (
        <Box sx={{ display: 'flex', justifyContent: 'center' }}>
            <Box sx={{ width: '100%' }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Typography variant="h4" gutterBottom>
                        Agent: {agentId}
                    </Typography>
                    <FormControl size="small" sx={{ minWidth: '12rem' }}>
                        <InputLabel id="range-label">Range</InputLabel>
                        <Select
                            labelId="range-label"
                            value={range}
                            label="Range"
                            onChange={handleRangeChange}
                        >
                            <MenuItem value="5m">Last 5 minutes</MenuItem>
                            <MenuItem value="15m">Last 15 minutes</MenuItem>
                            <MenuItem value="30m">Last 30 minutes</MenuItem>
                            <MenuItem value="1h">Last 1 hour</MenuItem>
                            <MenuItem value="6h">Last 6 hours</MenuItem>
                            <MenuItem value="24h">Last 24 hours</MenuItem>
                            <MenuItem value="7d">Last 7 days</MenuItem>
                        </Select>
                    </FormControl>
                </Box>

                {latest && (
                    <Box sx={{ display: 'flex', direction: 'row', gap: 2, mb: 2, flexWrap: 'nowrap', justifyContent: 'center' }}>
                        <Card sx={{ width: '100%' }}>
                            <CardContent sx={{ display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
                                <Typography color="text.secondary" gutterBottom>
                                    CPU Usage
                                </Typography>
                                <Typography variant="h4">
                                    {latest.cpu_percent.toFixed(1)}%
                                </Typography>
                                <Typography variant="body2" color="text.secondary">
                                    Cores: {latest.cpu_cores}
                                </Typography>
                            </CardContent>
                        </Card>
                        <Card sx={{ width: '100%' }}>
                            <CardContent sx={{ display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
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
                        <Card sx={{ width: '100%' }}>
                            <CardContent sx={{ display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
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
                        <Card sx={{ width: '100%' }}>
                            <CardContent sx={{ display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
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
                    </Box>
                )}

                {metrics.length > 0 && (
                    <Box sx={{
                        display: 'flex', flex: {
                            xs: "1 1 100%",
                            sm: "1 1 calc(50% - 1rem)",
                        }, gap: 2, mb: 2, flexWrap: 'wrap', justifyContent: 'center'
                    }}>
                        <Card sx={{ flex: "1 1 calc(50% - 1rem)", minWidth: 320 }}>
                            <CardContent>
                                <Typography variant="h6" gutterBottom>
                                    CPU Usage (%)
                                </Typography>
                                <Box sx={{ width: '100%', '& svg circle': { r: 0, display: 'none' }, '& svg path': { strokeWidth: 1.2 }, '& svg text': { fontSize: '0.85rem' } }}>
                                    <LineChart
                                        xAxis={[{ data: timestamps, scaleType: 'time', tickFormat: (d) => formatTick(new Date(d)) }]}
                                        series={[{ data: cpuData, label: 'CPU %', curve: 'linear' }]}
                                        yAxis={[{ min: 0, ...(cpuMax ? { max: cpuMax } : {}) }]}
                                        tooltip={{ xFormatter: (d) => formatTick(new Date(d)) }}
                                        height={240}
                                    />
                                </Box>
                            </CardContent>
                        </Card>
                        <Card sx={{ flex: "1 1 calc(50% - 1rem)", minWidth: 320 }}>
                            <CardContent>
                                <Typography variant="h6" gutterBottom>
                                    Memory Usage (%)
                                </Typography>
                                <Box sx={{ width: '100%', '& svg circle': { r: 0, display: 'none' }, '& svg path': { strokeWidth: 1.2 }, '& svg text': { fontSize: '0.85rem' } }}>
                                    <LineChart
                                        xAxis={[{ data: timestamps, scaleType: 'time', tickFormat: (d) => formatTick(new Date(d)) }]}
                                        series={[{ data: memoryData, label: 'Memory %', curve: 'linear' }]}
                                        yAxis={[{ min: 0, ...(memMax ? { max: memMax } : {}) }]}
                                        tooltip={{ xFormatter: (d) => formatTick(new Date(d)) }}
                                        height={240}
                                    />
                                </Box>
                            </CardContent>
                        </Card>
                        <Card sx={{ flex: "1 1 calc(50% - 1rem)", minWidth: 320 }}>
                            <CardContent>
                                <Typography variant="h6" gutterBottom>
                                    Disk Usage (%)
                                </Typography>
                                <Box sx={{ width: '100%', '& svg circle': { r: 0, display: 'none' }, '& svg path': { strokeWidth: 1.2 }, '& svg text': { fontSize: '0.85rem' } }}>
                                    <LineChart
                                        xAxis={[{ data: timestamps, scaleType: 'time', tickFormat: (d) => formatTick(new Date(d)) }]}
                                        series={[{ data: diskData, label: 'Disk %', curve: 'linear' }]}
                                        yAxis={[{ min: 0, ...(diskMax ? { max: diskMax } : {}) }]}
                                        tooltip={{ xFormatter: (d) => formatTick(new Date(d)) }}
                                        height={240}
                                    />
                                </Box>
                            </CardContent>
                        </Card>
                        <Card sx={{ flex: "1 1 calc(50% - 1rem)", minWidth: 320 }}>
                            <CardContent>
                                <Typography variant="h6" gutterBottom>
                                    Load Average (1m)
                                </Typography>
                                <Box sx={{ width: '100%', '& svg circle': { r: 0, display: 'none' }, '& svg path': { strokeWidth: 1.2 }, '& svg text': { fontSize: '0.85rem' } }}>
                                    <LineChart
                                        xAxis={[{ data: timestamps, scaleType: 'time', tickFormat: (d) => formatTick(new Date(d)) }]}
                                        series={[{ data: loadData, label: 'Load', curve: 'linear' }]}
                                        yAxis={[{ min: 0, ...(loadMax ? { max: loadMax } : {}) }]}
                                        tooltip={{ xFormatter: (d) => formatTick(new Date(d)) }}
                                        height={240}
                                    />
                                </Box>
                            </CardContent>
                        </Card>
                    </Box>
                )}
            </Box>
        </Box>
    );
}

export default AgentDetails;
