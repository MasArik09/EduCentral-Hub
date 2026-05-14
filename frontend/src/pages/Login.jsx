import { Eye, EyeOff, Lock, Mail, Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import api from "../api/axiosConfig";
import { useAuthStore } from "../store/authStore";

const getInitialTheme = () => {
  const saved = localStorage.getItem("theme");
  if (saved) {
    return saved === "dark";
  }

  return window.matchMedia?.("(prefers-color-scheme: dark)").matches ?? false;
};

function Login() {
  const navigate = useNavigate();
  const setToken = useAuthStore((state) => state.setToken);
  const setUser = useAuthStore((state) => state.setUser);
  const [isDark, setIsDark] = useState(getInitialTheme);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    const root = document.documentElement;
    root.classList.toggle("dark", isDark);
    localStorage.setItem("theme", isDark ? "dark" : "light");
  }, [isDark]);

  const subtitle =
    "Streamline assignments, attendance, and lessons in one unified platform.";

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");
    setIsSubmitting(true);

    try {
      const response = await api.post("/auth/login", {
        email,
        password,
      });
      const { token, user } = response.data || {};

      if (token) {
        setToken(token);
      }
      if (user) {
        setUser(user);
      }

      navigate("/dashboard");
    } catch (err) {
      const message =
        err?.response?.data?.message ||
        "Login failed. Please check your credentials.";
      setError(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen w-full bg-slate-50 text-slate-900 dark:bg-[#0b1220] dark:text-slate-100">
      <div className="relative min-h-screen w-full">
        <div className="absolute right-6 top-6 z-20">
          <button
            type="button"
            onClick={() => setIsDark((value) => !value)}
            className="inline-flex items-center gap-2 rounded-full border border-slate-200/70 bg-white/80 px-4 py-2 text-sm font-medium text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-white dark:border-slate-700 dark:bg-slate-900/80 dark:text-slate-200"
          >
            {isDark ? (
              <Sun className="h-4 w-4" />
            ) : (
              <Moon className="h-4 w-4" />
            )}
            {isDark ? "Light" : "Dark"} mode
          </button>
        </div>

        <div className="flex min-h-screen w-full flex-col lg:flex-row">
          <div className="relative flex w-full flex-col justify-between overflow-hidden bg-gradient-to-br from-sky-600 via-blue-600 to-indigo-700 px-8 py-12 text-white lg:w-[57%] lg:px-14">
            <div className="absolute -left-20 -top-20 h-64 w-64 rounded-full bg-white/10 blur-3xl" />
            <div className="absolute -bottom-28 right-10 h-72 w-72 rounded-full bg-indigo-400/30 blur-3xl" />
            <div className="absolute right-0 top-1/2 h-40 w-40 -translate-y-1/2 rounded-full bg-cyan-300/30 blur-2xl" />

            <div className="relative z-10">
              <span className="inline-flex items-center gap-2 rounded-full bg-white/15 px-4 py-1 text-xs font-semibold uppercase tracking-[0.2em]">
                EduCentral Hub
              </span>
              <h1 className="mt-8 text-5xl font-bold leading-tight md:text-6xl">
                Teach smarter.
                <br />
                Manage everything.
              </h1>
              <p className="mt-4 max-w-md text-base text-white/80 md:text-lg">
                {subtitle}
              </p>
            </div>

            <div className="relative z-10 mt-12 grid gap-6 text-sm text-white/80 md:grid-cols-2">
              <div className="rounded-2xl border border-white/30 bg-white/15 p-4 shadow-lg shadow-blue-900/30 backdrop-blur-xl">
                <p className="text-2xl font-semibold text-white">98%</p>
                <p className="mt-1 text-xs uppercase tracking-[0.2em]">
                  On-time submissions
                </p>
              </div>
              <div className="rounded-2xl border border-white/30 bg-white/15 p-4 shadow-lg shadow-blue-900/30 backdrop-blur-xl">
                <p className="text-2xl font-semibold text-white">24/7</p>
                <p className="mt-1 text-xs uppercase tracking-[0.2em]">
                  Smart monitoring
                </p>
              </div>
            </div>

            <div className="relative z-10 mt-10 text-xs uppercase tracking-[0.25em] text-white/60">
              Built for modern campuses
            </div>
          </div>

          <div className="flex w-full items-center justify-center bg-slate-950 px-6 py-12 text-slate-100 lg:w-[43%] lg:px-12">
            <div className="w-full max-w-md">
              <div className="rounded-3xl border border-slate-200/70 bg-white/80 p-8 shadow-xl shadow-slate-200/50 backdrop-blur transition dark:border-slate-800 dark:bg-slate-900/70 dark:shadow-none">
                <div className="flex items-center gap-3">
                  <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-sky-100 text-sky-600 dark:bg-slate-800 dark:text-sky-300">
                    <Lock className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="text-sm font-semibold uppercase tracking-[0.3em] text-sky-600 dark:text-sky-300">
                      Welcome back
                    </p>
                    <h2 className="text-2xl font-semibold text-slate-900 dark:text-white">
                      Sign in to your account
                    </h2>
                  </div>
                </div>

                <form onSubmit={handleSubmit} className="mt-8 space-y-5">
                  <div>
                    <label className="text-sm font-medium text-slate-600 dark:text-slate-300">
                      Email address
                    </label>
                    <div className="relative mt-2">
                      <Mail className="pointer-events-none absolute left-3 top-3.5 h-5 w-5 text-slate-400" />
                      <input
                        type="email"
                        value={email}
                        onChange={(event) => setEmail(event.target.value)}
                        placeholder="you@school.edu"
                        autoComplete="email"
                        required
                        className="h-12 w-full rounded-2xl border border-slate-200 bg-white px-10 text-sm text-slate-800 shadow-sm outline-none transition focus:border-sky-400 focus:ring-4 focus:ring-sky-100 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100 dark:focus:border-sky-400 dark:focus:ring-sky-900"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="text-sm font-medium text-slate-600 dark:text-slate-300">
                      Password
                    </label>
                    <div className="relative mt-2">
                      <Lock className="pointer-events-none absolute left-3 top-3.5 h-5 w-5 text-slate-400" />
                      <input
                        type={showPassword ? "text" : "password"}
                        value={password}
                        onChange={(event) => setPassword(event.target.value)}
                        placeholder="Enter your password"
                        autoComplete="current-password"
                        required
                        className="h-12 w-full rounded-2xl border border-slate-200 bg-white px-10 pr-12 text-sm text-slate-800 shadow-sm outline-none transition focus:border-sky-400 focus:ring-4 focus:ring-sky-100 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-100 dark:focus:border-sky-400 dark:focus:ring-sky-900"
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword((value) => !value)}
                        className="absolute right-3 top-3.5 text-slate-400 transition hover:text-slate-600 dark:hover:text-slate-200"
                        aria-label={
                          showPassword ? "Hide password" : "Show password"
                        }
                      >
                        {showPassword ? (
                          <EyeOff className="h-5 w-5" />
                        ) : (
                          <Eye className="h-5 w-5" />
                        )}
                      </button>
                    </div>
                  </div>

                  <div className="flex items-center justify-between text-sm">
                    <label className="flex items-center gap-2 text-slate-500 dark:text-slate-400">
                      <input
                        type="checkbox"
                        className="h-4 w-4 rounded border-slate-300 text-sky-600 focus:ring-sky-200 dark:border-slate-700 dark:focus:ring-sky-900"
                      />
                      Remember me
                    </label>
                    <button
                      type="button"
                      className="font-medium text-sky-600 transition hover:text-sky-700 dark:text-sky-300"
                    >
                      Forgot password?
                    </button>
                  </div>

                  {error ? (
                    <div className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-600 dark:border-rose-500/40 dark:bg-rose-500/10 dark:text-rose-200">
                      {error}
                    </div>
                  ) : null}

                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="flex h-12 w-full items-center justify-center gap-2 rounded-2xl bg-sky-600 text-sm font-semibold text-white shadow-lg shadow-sky-200/80 transition hover:bg-sky-700 focus:outline-none focus:ring-4 focus:ring-sky-200 disabled:cursor-not-allowed disabled:opacity-70 dark:shadow-none dark:focus:ring-sky-900"
                  >
                    {isSubmitting ? "Signing in..." : "Sign in"}
                  </button>
                </form>

                <p className="mt-6 text-center text-xs text-slate-500 dark:text-slate-400">
                  By signing in, you agree to our Terms and Privacy Policy.
                </p>
              </div>

              <p className="mt-6 text-center text-xs text-slate-500 dark:text-slate-400">
                New to EduCentral Hub? Contact your administrator to onboard.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Login;
