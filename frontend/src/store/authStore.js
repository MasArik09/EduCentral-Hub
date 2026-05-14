import { create } from "zustand";

const loadToken = () => localStorage.getItem("token");
const loadUser = () => {
  const raw = localStorage.getItem("user");
  return raw ? JSON.parse(raw) : null;
};

export const useAuthStore = create((set) => ({
  token: loadToken(),
  user: loadUser(),
  setToken: (token) => {
    if (token) {
      localStorage.setItem("token", token);
    } else {
      localStorage.removeItem("token");
    }

    set({ token });
  },
  setUser: (user) => {
    if (user) {
      localStorage.setItem("user", JSON.stringify(user));
    } else {
      localStorage.removeItem("user");
    }

    set({ user });
  },
  clearAuth: () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    set({ token: null, user: null });
  },
}));
