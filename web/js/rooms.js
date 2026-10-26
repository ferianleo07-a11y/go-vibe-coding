// rooms.js - Manajemen Kamar & Ketersediaan Real-Time

let currentRooms = [];

// Muat data kamar dari backend
async function loadRooms(statusFilter = "") {
  try {
    let url = "/api/rooms";
    if (statusFilter) {
      url += `?status=${encodeURIComponent(statusFilter)}`;
    }

    const res = await authFetch(url);
    if (!res.ok) {
      const errData = await res.json();
      throw new Error(errData.error || "Gagal memuat kamar");
    }

    const result = await res.json();
    currentRooms = result.data || [];
    renderRooms(currentRooms);
  } catch (err) {
    console.error("Gagal memuat kamar:", err);
    const tbody = document.getElementById("roomsTableBody");
    if (tbody) {
      tbody.innerHTML = `<tr><td colspan="6" style="text-align: center; color: var(--danger); padding: 1.5rem;">${err.message}</td></tr>`;
    }
  }
}

// Render tabel kamar
function renderRooms(rooms) {
  const tbody = document.getElementById("roomsTableBody");
  if (!tbody) return;

  if (rooms.length === 0) {
    tbody.innerHTML = `
      <tr>
        <td colspan="6" style="text-align: center; color: var(--text-muted); padding: 2rem;">
          Belum ada data kamar. Klik tombol "Tambah Kamar" untuk memulai.
        </td>
      </tr>
    `;
    return;
  }

  tbody.innerHTML = rooms.map(room => {
    const badgeClass = getStatusBadgeClass(room.status);
    const formattedPrice = new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(room.price_per_night);

    return `
      <tr>
        <td><strong>${escapeHtml(room.room_number)}</strong></td>
        <td>${escapeHtml(room.type)}</td>
        <td>${formattedPrice}</td>
        <td><span class="badge ${badgeClass}">${escapeHtml(room.status)}</span></td>
        <td><small style="color: var(--text-muted);">${escapeHtml(room.description || "-")}</small></td>
        <td>
          <button class="btn-sm btn-outline" onclick="openEditRoomModal(${room.id})">Edit</button>
          <button class="btn-sm btn-danger" onclick="deleteRoom(${room.id}, '${escapeHtml(room.room_number)}')">Hapus</button>
        </td>
      </tr>
    `;
  }).join("");
}

function getStatusBadgeClass(status) {
  switch (status) {
    case "Tersedia": return "badge-success";
    case "Terisi": return "badge-primary";
    case "Sedang Dibersihkan": return "badge-warning";
    case "Dalam Perbaikan": return "badge-danger";
    default: return "badge-secondary";
  }
}

function escapeHtml(str) {
  if (!str) return "";
  return String(str)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

// Buka modal tambah kamar baru
function openAddRoomModal() {
  document.getElementById("roomModalTitle").textContent = "Tambah Kamar Baru";
  document.getElementById("roomId").value = "";
  document.getElementById("roomNumber").value = "";
  document.getElementById("roomType").value = "Standard";
  document.getElementById("roomPrice").value = "";
  document.getElementById("roomStatus").value = "Tersedia";
  document.getElementById("roomDescription").value = "";
  document.getElementById("roomModalAlert").style.display = "none";
  document.getElementById("roomModal").style.display = "flex";
}

// Buka modal edit kamar
function openEditRoomModal(id) {
  const room = currentRooms.find(r => r.id === id);
  if (!room) return;

  document.getElementById("roomModalTitle").textContent = `Edit Kamar ${room.room_number}`;
  document.getElementById("roomId").value = room.id;
  document.getElementById("roomNumber").value = room.room_number;
  document.getElementById("roomType").value = room.type;
  document.getElementById("roomPrice").value = room.price_per_night;
  document.getElementById("roomStatus").value = room.status;
  document.getElementById("roomDescription").value = room.description || "";
  document.getElementById("roomModalAlert").style.display = "none";
  document.getElementById("roomModal").style.display = "flex";
}

// Tutup modal
function closeRoomModal() {
  document.getElementById("roomModal").style.display = "none";
}

// Simpan data kamar (POST atau PUT)
async function handleRoomFormSubmit(e) {
  e.preventDefault();
  const alertEl = document.getElementById("roomModalAlert");
  alertEl.style.display = "none";

  const id = document.getElementById("roomId").value;
  const payload = {
    room_number: document.getElementById("roomNumber").value.trim(),
    type: document.getElementById("roomType").value.trim(),
    price_per_night: parseFloat(document.getElementById("roomPrice").value),
    status: document.getElementById("roomStatus").value,
    description: document.getElementById("roomDescription").value.trim()
  };

  try {
    const isEdit = !!id;
    const url = isEdit ? `/api/rooms/${id}` : "/api/rooms";
    const method = isEdit ? "PUT" : "POST";

    const res = await authFetch(url, {
      method: method,
      body: JSON.stringify(payload)
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || "Gagal menyimpan data kamar");
    }

    closeRoomModal();
    loadRooms();
  } catch (err) {
    alertEl.textContent = err.message;
    alertEl.style.display = "block";
  }
}

// Hapus kamar
async function deleteRoom(id, roomNumber) {
  if (!confirm(`Apakah Anda yakin ingin menghapus kamar ${roomNumber}?`)) {
    return;
  }

  try {
    const res = await authFetch(`/api/rooms/${id}`, {
      method: "DELETE"
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || "Gagal menghapus kamar");
    }

    loadRooms();
  } catch (err) {
    alert("Error: " + err.message);
  }
}
