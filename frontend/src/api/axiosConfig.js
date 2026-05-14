import axios from "axios";
import { useAuthStore } from "../store/authStore";

const api = axios.create({
  baseURL: "http://localhost:8080/api",
  withCredentials: true,
});

api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token || localStorage.getItem("token");

    if (token) {
      config.headers = config.headers || {};
      config.headers.Authorization = `Bearer ${token}`;
    }

    return config;
  },
  (error) => Promise.reject(error)
);

api.interceptors.response.use(
  (response) => {
    const accessToken = response.headers["x-access-token"];

    if (accessToken) {
      const { setToken } = useAuthStore.getState();
      setToken(accessToken);
    }

    return response;
  },
  (error) => Promise.reject(error)
);

export default api;
