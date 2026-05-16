import React, { useState } from "react";

const Login = () => {
  const [showPassword, setShowPassword] = useState(false);

  return (
    <div className="w-full min-h-screen bg-[#090D16] flex flex-col lg:flex-row overflow-hidden font-sans">
      <div
        style={{ width: "90%" }}
        className="hidden lg:flex bg-gradient-to-b from-[#1E40AF] via-[#0369A1] to-[#0D9488] p-12 flex-col justify-between"
      >
        <div>
          <span className="text-xs font-medium tracking-widest text-cyan-200 uppercase bg-white/10 px-3 py-1.5 rounded-md backdrop-blur-sm">
            Portal Hub
          </span>
          <h1 className="text-4xl font-extrabold text-white mt-10 tracking-tight leading-tight">
            EduCentral
            <br />
            Management.
          </h1>
          <p className="text-white/70 text-sm mt-4 max-w-xs leading-relaxed">
            Satu akses terintegrasi untuk seluruh kegiatan akademik, kelas, dan
            manajemen data esensial.
          </p>
        </div>

        <div className="text-white/40 text-xs font-mono tracking-widest uppercase">
          v1.0.0 // Enterprise Secure
        </div>
      </div>

      <div
        style={{ width: "70%" }}
        className="w-full min-h-screen flex flex-col justify-center items-center bg-[#090D16] p-6"
      >
        <div className="w-full max-w-md bg-[#111827] p-10 rounded-2xl border border-slate-800/80 shadow-2xl">
          <div className="mb-8">
            <h2 className="text-2xl font-bold text-slate-100 tracking-tight">
              Selamat Datang
            </h2>
            <p className="text-slate-400 text-sm mt-1">
              Silakan masuk menggunakan akun institusi Anda.
            </p>
          </div>

          <form className="space-y-5" onSubmit={(e) => e.preventDefault()}>
            <div>
              <label className="text-xs font-medium text-slate-300 block mb-2 tracking-wide">
                Email Institusi
              </label>
              <input
                type="email"
                placeholder="nama@university.edu"
                className="w-full bg-[#1F2937] text-slate-100 px-4 py-3 rounded-xl border border-slate-700/60 focus:outline-none focus:border-cyan-500 text-sm transition-all placeholder:text-slate-500"
              />
            </div>

            <div>
              <label className="text-xs font-medium text-slate-300 block mb-2 tracking-wide">
                Kata Sandi
              </label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  placeholder="Masukkan kata sandi Anda"
                  className="w-full bg-[#1F2937] text-slate-100 px-4 py-3 rounded-xl border border-slate-700/60 focus:outline-none focus:border-cyan-500 text-sm transition-all placeholder:text-slate-500 pr-16"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-xs font-medium text-cyan-400 hover:text-cyan-300 transition-colors"
                >
                  {showPassword ? "Sembunyikan" : "Lihat"}
                </button>
              </div>
            </div>

            <div className="flex items-center justify-between text-xs pt-1">
              <label className="flex items-center space-x-2 text-slate-400 cursor-pointer select-none">
                <input
                  type="checkbox"
                  className="rounded bg-[#1F2937] border-slate-700 text-cyan-500 focus:ring-0 w-4 h-4"
                />
                <span>Ingat perangkat ini</span>
              </label>
              <a href="#" className="text-cyan-400 hover:underline transition-all">
                Lupa sandi?
              </a>
            </div>

            <button
              type="submit"
              className="w-full bg-blue-600 hover:bg-blue-500 text-white font-semibold py-3.5 rounded-xl transition-all text-sm shadow-lg shadow-blue-600/10 mt-6 tracking-wide"
            >
              Masuk via SSO
            </button>
          </form>

        </div>

        <div className="text-xs text-slate-500 mt-8">
          Belum terdaftar?{" "}
          <span className="text-cyan-400 cursor-pointer hover:underline">
            Hubungi Admin
          </span>
        </div>
      </div>
    </div>
  );
};

export default Login;
