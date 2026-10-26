// auth.js - Pengelolaan Autentikasi dan Token JWT di localStorage

const TOKEN_KEY = "hotel_auth_token";
const USER_KEY = "hotel_auth_user";

// Simpan token & profil pengguna di localStorage
function setAuthSession(token, user) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

// Ambil token JWT dari localStorage
function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

// Ambil data profil pengguna dari localStorage
function getCurrentUser() {
  const user = localStorage.getItem(USER_KEY);
  return user ? JSON.parse(user) : null;
}

// Hapus sesi login saat logout
function logout() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
  window.location.href = "/login";
}

// Helper untuk fetch API dengan header Authorization Bearer
async function authFetch(url, options = {}) {
  const token = getToken();
  const headers = {
    "Content-Type": "application/json",
    ...(token ? { "Authorization": `Bearer ${token}` } : {}),
    ...(options.headers || {})
  };

  const response = await fetch(url, { ...options, headers });
  if (response.status === 401) {
    // Sesi token tidak valid atau kedaluwarsa
    logout();
    throw new Error("Sesi login berakhir. Silakan masuk kembali.");
  }
  return response;
}

// Inisialisasi form Registrasi (di halaman register.html)
function initRegisterForm() {
  const form = document.getElementById("registerForm");
  const alertEl = document.getElementById("alertMessage");

  if (!form) return;

  // Jika sudah login, redirect ke halaman utama
  if (getToken()) {
    window.location.href = "/";
    return;
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    showAlert("", false);

    const name = document.getElementById("name").value.trim();
    const email = document.getElementById("email").value.trim();
    const password = document.getElementById("password").value;

    try {
      const res = await fetch("/api/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, email, password })
      });

      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || "Gagal melakukan pendaftaran");
      }

      showAlert("Pendaftaran berhasil! Mengalihkan ke halaman login...", true);
      setTimeout(() => {
        window.location.href = "/login";
      }, 1500);
    } catch (err) {
      showAlert(err.message, false);
    }
  });
}

// Inisialisasi form Login (di halaman login.html)
function initLoginForm() {
  const form = document.getElementById("loginForm");
  const alertEl = document.getElementById("alertMessage");

  if (!form) return;

  // Jika sudah login, redirect ke halaman utama
  if (getToken()) {
    window.location.href = "/";
    return;
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    showAlert("", false);

    const email = document.getElementById("email").value.trim();
    const password = document.getElementById("password").value;

    try {
      const res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password })
      });

      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || "Gagal masuk ke akun");
      }

      // Simpan JWT di localStorage
      setAuthSession(data.token, data.user);
      showAlert("Login berhasil! Mengalihkan...", true);
      setTimeout(() => {
        window.location.href = "/";
      }, 800);
    } catch (err) {
      showAlert(err.message, false);
    }
  });
}

function showAlert(message, isSuccess = false) {
  const alertEl = document.getElementById("alertMessage");
  if (!alertEl) return;

  if (!message) {
    alertEl.style.display = "none";
    return;
  }

  alertEl.textContent = message;
  alertEl.className = isSuccess ? "alert alert-success" : "alert alert-danger";
  alertEl.style.display = "block";
}
