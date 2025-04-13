// src/main.tsx
import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import "./index.css";
import axios from "axios";

// Configure axios defaults
axios.defaults.baseURL = window.location.origin;
axios.defaults.headers.common["Content-Type"] = "application/json";

// Set up interceptor for token refresh
axios.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // If the error is 401 and we haven't already tried to refresh
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = localStorage.getItem("refreshToken");

        if (refreshToken) {
          // Attempt to refresh the token
          const response = await axios.post("/api/auth/refresh", {
            refresh_token: refreshToken,
          });

          if (response.data.success) {
            // Update the access token
            const newAccessToken = response.data.data.access_token;
            localStorage.setItem("accessToken", newAccessToken);

            // Update the header and retry the original request
            axios.defaults.headers.common[
              "Authorization"
            ] = `Bearer ${newAccessToken}`;
            originalRequest.headers[
              "Authorization"
            ] = `Bearer ${newAccessToken}`;

            return axios(originalRequest);
          }
        }
      } catch (refreshError) {
        console.error("Failed to refresh token:", refreshError);
      }

      // If we get here, we couldn't refresh the token
      localStorage.removeItem("accessToken");
      localStorage.removeItem("refreshToken");
      window.location.href = "/login";
    }

    return Promise.reject(error);
  }
);

// Initialize auth header from stored token
const token = localStorage.getItem("accessToken");
if (token) {
  axios.defaults.headers.common["Authorization"] = `Bearer ${token}`;
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
