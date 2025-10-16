import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import Login from './components/Login';
import Dashboard from './components/Dashboard';
import Install from './components/Install';
import axios from 'axios';

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
  },
});

function App() {
  const [token, setToken] = useState(localStorage.getItem('token'));
  const [installed, setInstalled] = useState(null);
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    checkInstallation();
  }, []);

  const checkInstallation = async () => {
    try {
      const response = await axios.get('/api/install/check');
      setInstalled(response.data.installed);
    } catch (error) {
      console.error('Failed to check installation:', error);
      setInstalled(true); // Assume installed if check fails
    } finally {
      setChecking(false);
    }
  };

  const handleLogin = (newToken) => {
    localStorage.setItem('token', newToken);
    setToken(newToken);
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    setToken(null);
  };

  const handleInstalled = () => {
    setInstalled(true);
  };

  if (checking) {
    return (
      <ThemeProvider theme={theme}>
        <CssBaseline />
      </ThemeProvider>
    );
  }

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Router>
        {!installed ? (
          <Install onInstalled={handleInstalled} />
        ) : (
          <Routes>
            <Route 
              path="/login" 
              element={token ? <Navigate to="/" /> : <Login onLogin={handleLogin} />} 
            />
            <Route 
              path="/*" 
              element={token ? <Dashboard onLogout={handleLogout} token={token} /> : <Navigate to="/login" />} 
            />
          </Routes>
        )}
      </Router>
    </ThemeProvider>
  );
}

export default App;
