// reservations.js - Check-in, Check-out, Kalkulasi Tagihan, Metode Pembayaran & Reminder WA/Email

let currentReservations = [];

// Muat data reservasi dari server
async function loadReservations(statusFilter = "") {
  try {
    let url = "/api/reservations";
    if (statusFilter) {
      url += `?status=${encodeURIComponent(statusFilter)}`;
    }

    const res = await authFetch(url);
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || "Gagal memuat transaksi reservasi");
    }

    const result = await res.json();
    currentReservations = result.data || [];
    renderReservations(currentReservations);
  } catch (err) {
    console.error("Gagal memuat reservasi:", err);
    const tbody = document.getElementById("reservationsTableBody");
    if (tbody) {
      tbody.innerHTML = `<tr><td colspan="9" style="text-align: center; color: var(--danger); padding: 1.5rem;">${err.message}</td></tr>`;
    }
  }
}

// Render data reservasi pada tabel
function renderReservations(reservations) {
  const tbody = document.getElementById("reservationsTableBody");
  if (!tbody) return;

  if (reservations.length === 0) {
    tbody.innerHTML = `
      <tr>
        <td colspan="9" style="text-align: center; color: var(--text-muted); padding: 2rem;">
          Belum ada data transaksi tamu. Klik tombol "Check-in Tamu" untuk memulai.
        </td>
      </tr>
    `;
    return;
  }

  const formatIDR = (num) => new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(num);

  tbody.innerHTML = reservations.map(r => {
    const statusBadge = getReservationBadge(r.status);
    const payBadge = getPaymentBadge(r.payment_status);
    const reminderBadge = getReminderBadge(r.reminder_status);
    const inDate = r.check_in_date.substring(0, 10);
    const outDate = r.check_out_date.substring(0, 10);
    const roomNumber = r.room ? r.room.room_number : `ID #${r.room_id}`;
    const roomType = r.room ? r.room.type : "";
    const paymentMethod = r.payment_method || "Tunai";

    return `
      <tr>
        <td><strong>${escapeHtml(r.invoice_number)}</strong></td>
        <td>
          <div><strong>${escapeHtml(r.guest_name)}</strong></div>
          <small style="color: var(--text-muted);">${escapeHtml(r.guest_phone || "-")}</small>
          ${r.guest_email ? `<div style="font-size: 0.75rem; color: var(--text-muted);">${escapeHtml(r.guest_email)}</div>` : ''}
        </td>
        <td>
          <div><strong>Kamar ${escapeHtml(roomNumber)}</strong></div>
          <small style="color: var(--text-muted);">${escapeHtml(roomType)}</small>
        </td>
        <td>
          <div>${inDate} s/d ${outDate}</div>
          <small style="color: var(--primary); font-weight: 600;">${r.nights} Malam</small>
        </td>
        <td>
          <div>${formatIDR(r.total_amount)}</div>
          <div style="margin-top: 0.2rem;">
            <span class="badge badge-secondary" style="font-size: 0.7rem;">💳 ${escapeHtml(paymentMethod)}</span>
          </div>
          <small style="color: var(--text-muted);">Sisa: ${formatIDR(r.remaining_amount)}</small>
        </td>
        <td><span class="badge ${payBadge}">${escapeHtml(r.payment_status)}</span></td>
        <td><span class="badge ${statusBadge}">${escapeHtml(r.status)}</span></td>
        <td>
          <span class="badge ${reminderBadge.class}">${reminderBadge.label}</span>
          ${r.reminder_sent_at ? `<div style="font-size: 0.7rem; color: var(--text-muted); margin-top: 0.2rem;">${r.reminder_sent_at.substring(11, 16)} WIB</div>` : ''}
        </td>
        <td>
          <div style="display: flex; flex-wrap: wrap; gap: 0.25rem;">
            <button class="btn-sm btn-outline" onclick="openInvoiceModal(${r.id})">🧾 Nota</button>
            ${r.status === 'Check-in' ? `
              <button class="btn-sm btn-outline" style="color: #047857; border-color: #059669;" onclick="sendReminder(${r.id}, '${escapeHtml(r.guest_name)}')">🔔 Remind</button>
              <button class="btn-sm btn-primary" onclick="checkoutGuest(${r.id}, '${escapeHtml(r.guest_name)}')">Check-out</button>
              <button class="btn-sm btn-danger" onclick="cancelGuest(${r.id}, '${escapeHtml(r.guest_name)}')">Batal</button>
            ` : ''}
          </div>
        </td>
      </tr>
    `;
  }).join("");
}

function getReservationBadge(status) {
  switch (status) {
    case "Check-in": return "badge-primary";
    case "Check-out": return "badge-success";
    case "Dibatalkan": return "badge-danger";
    default: return "badge-secondary";
  }
}

function getPaymentBadge(status) {
  switch (status) {
    case "Lunas": return "badge-success";
    case "DP": return "badge-warning";
    case "Belum Bayar": return "badge-danger";
    default: return "badge-secondary";
  }
}

function getReminderBadge(status) {
  switch (status) {
    case "Terkirim":
      return { class: "badge-success", label: "🟢 Terkirim" };
    case "Gagal":
      return { class: "badge-danger", label: "🔴 Gagal" };
    default:
      return { class: "badge-secondary", label: "⚪ Belum Dikirim" };
  }
}

// Buka modal Check-in baru
async function openCheckInModal() {
  const alertEl = document.getElementById("checkInModalAlert");
  alertEl.style.display = "none";

  // Ambil hanya kamar dengan status Tersedia
  const selectRoom = document.getElementById("resRoomId");
  selectRoom.innerHTML = `<option value="">Memuat kamar tersedia...</option>`;

  try {
    const res = await authFetch("/api/rooms?status=Tersedia");
    const data = await res.json();
    const availableRooms = data.data || [];

    if (availableRooms.length === 0) {
      selectRoom.innerHTML = `<option value="">Tidak ada kamar tersedia saat ini</option>`;
    } else {
      selectRoom.innerHTML = availableRooms.map(r => 
        `<option value="${r.id}" data-price="${r.price_per_night}">
          Kamar ${escapeHtml(r.room_number)} (${escapeHtml(r.type)}) - Rp ${r.price_per_night.toLocaleString('id-ID')}/malam
        </option>`
      ).join("");
    }
  } catch (err) {
    selectRoom.innerHTML = `<option value="">Gagal memuat kamar</option>`;
  }

  // Set tanggal default: hari ini dan besok
  const today = new Date();
  const tomorrow = new Date();
  tomorrow.setDate(today.getDate() + 1);

  document.getElementById("resCheckInDate").value = today.toISOString().split("T")[0];
  document.getElementById("resCheckOutDate").value = tomorrow.toISOString().split("T")[0];
  document.getElementById("resGuestName").value = "";
  document.getElementById("resGuestID").value = "";
  document.getElementById("resGuestPhone").value = "";
  const emailInput = document.getElementById("resGuestEmail");
  if (emailInput) emailInput.value = "";
  document.getElementById("resDownPayment").value = "0";
  const payMethodInput = document.getElementById("resPaymentMethod");
  if (payMethodInput) payMethodInput.value = "Tunai";
  document.getElementById("resNotes").value = "";

  calculateEstimate();
  document.getElementById("checkInModal").style.display = "flex";
}

function closeCheckInModal() {
  document.getElementById("checkInModal").style.display = "none";
}

// Kalkulasi dinamis durasi malam & total estimasi
function calculateEstimate() {
  const selectRoom = document.getElementById("resRoomId");
  const selectedOption = selectRoom.options[selectRoom.selectedIndex];
  const price = selectedOption && selectedOption.dataset.price ? parseFloat(selectedOption.dataset.price) : 0;

  const inDateVal = document.getElementById("resCheckInDate").value;
  const outDateVal = document.getElementById("resCheckOutDate").value;
  const dpVal = parseFloat(document.getElementById("resDownPayment").value) || 0;

  if (!inDateVal || !outDateVal) return;

  const d1 = new Date(inDateVal);
  const d2 = new Date(outDateVal);
  const diffTime = d2.getTime() - d1.getTime();
  let nights = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  if (nights < 1) nights = 1;

  const total = nights * price;
  let remaining = total - dpVal;
  if (remaining < 0) remaining = 0;

  document.getElementById("estNights").textContent = `${nights} Malam`;
  document.getElementById("estTotal").textContent = `Rp ${total.toLocaleString('id-ID')}`;
  document.getElementById("estRemaining").textContent = `Rp ${remaining.toLocaleString('id-ID')}`;
}

// Submit Check-in form
async function handleCheckInSubmit(e) {
  e.preventDefault();
  const alertEl = document.getElementById("checkInModalAlert");
  alertEl.style.display = "none";

  const roomId = document.getElementById("resRoomId").value;
  if (!roomId) {
    alertEl.textContent = "Silakan pilih kamar yang tersedia";
    alertEl.style.display = "block";
    return;
  }

  const payload = {
    room_id: parseInt(roomId, 10),
    guest_name: document.getElementById("resGuestName").value.trim(),
    guest_id_card: document.getElementById("resGuestID").value.trim(),
    guest_phone: document.getElementById("resGuestPhone").value.trim(),
    guest_email: document.getElementById("resGuestEmail") ? document.getElementById("resGuestEmail").value.trim() : "",
    check_in_date: document.getElementById("resCheckInDate").value,
    check_out_date: document.getElementById("resCheckOutDate").value,
    down_payment: parseFloat(document.getElementById("resDownPayment").value) || 0,
    payment_method: document.getElementById("resPaymentMethod") ? document.getElementById("resPaymentMethod").value : "Tunai",
    notes: document.getElementById("resNotes").value.trim()
  };

  try {
    const res = await authFetch("/api/reservations", {
      method: "POST",
      body: JSON.stringify(payload)
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || "Gagal memproses check-in");
    }

    closeCheckInModal();
    loadReservations();
    // Refresh juga data kamar karena status kamar telah berubah menjadi Terisi
    if (typeof loadRooms === "function") {
      loadRooms();
    }
  } catch (err) {
    alertEl.textContent = err.message;
    alertEl.style.display = "block";
  }
}

// Kirim Pengingat Checkout (WA & Email)
async function sendReminder(id, guestName) {
  if (!confirm(`Kirim pengingat check-out (15 menit) via WhatsApp dan Email kepada "${guestName}"?`)) {
    return;
  }

  try {
    const res = await authFetch(`/api/reservations/${id}/remind`, {
      method: "POST"
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || "Gagal mengirimkan pengingat");
    }

    let alertMsg = `Pengingat berhasil diproses!\n- ${data.details || 'WA & Email terkirim'}`;
    if (data.whatsapp_url) {
      alertMsg += `\n\nKlik OK untuk membuka WhatsApp Web / Chat langsung dengan tamu.`;
      if (confirm(alertMsg)) {
        window.open(data.whatsapp_url, "_blank");
      }
    } else {
      alert(alertMsg);
    }

    loadReservations();
  } catch (err) {
    alert("Error: " + err.message);
  }
}

// Proses Check-out
async function checkoutGuest(id, guestName) {
  if (!confirm(`Konfirmasi Check-out untuk tamu "${guestName}"? Sisa pembayaran akan ditandai Lunas dan status kamar menjadi 'Sedang Dibersihkan'.`)) {
    return;
  }

  try {
    const res = await authFetch(`/api/reservations/${id}/checkout`, {
      method: "PUT"
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || "Gagal memproses check-out");
    }

    loadReservations();
    if (typeof loadRooms === "function") {
      loadRooms();
    }
  } catch (err) {
    alert("Error: " + err.message);
  }
}

// Proses Pembatalan
async function cancelGuest(id, guestName) {
  if (!confirm(`Apakah Anda yakin ingin membatalkan reservasi tamu "${guestName}"? Kamar akan dikembalikan berstatus 'Tersedia'.`)) {
    return;
  }

  try {
    const res = await authFetch(`/api/reservations/${id}/cancel`, {
      method: "PUT"
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || "Gagal membatalkan reservasi");
    }

    loadReservations();
    if (typeof loadRooms === "function") {
      loadRooms();
    }
  } catch (err) {
    alert("Error: " + err.message);
  }
}

// Buka Modal Cetak Nota / Invoice
async function openInvoiceModal(id) {
  try {
    const res = await authFetch(`/api/reservations/${id}`);
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "Gagal memuat detail nota");

    const r = data.data;
    const formatIDR = (num) => new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(num);

    const user = getCurrentUser() || { name: "Hotel Finance" };
    const paymentMethod = r.payment_method || "Tunai";

    const invoiceHTML = `
      <div id="printableInvoice" class="invoice-container">
        <div class="invoice-header">
          <div>
            <h2 style="color: var(--primary);">${escapeHtml(user.name)}</h2>
            <p style="color: var(--text-muted); font-size: 0.85rem;">Official Booking Invoice & Receipt</p>
          </div>
          <div style="text-align: right;">
            <h3 style="margin-bottom: 0.25rem;">${escapeHtml(r.invoice_number)}</h3>
            <p style="font-size: 0.85rem; color: var(--text-muted);">Tanggal: ${r.created_at.substring(0, 10)}</p>
          </div>
        </div>

        <hr style="border: none; border-top: 1px solid var(--border); margin: 1rem 0;">

        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-bottom: 1.5rem; font-size: 0.875rem;">
          <div>
            <strong>Informasi Tamu:</strong>
            <div>Nama: ${escapeHtml(r.guest_name)}</div>
            <div>No. Identitas (KTP/Passport): ${escapeHtml(r.guest_id_card)}</div>
            <div>Telepon / WhatsApp: ${escapeHtml(r.guest_phone || "-")}</div>
            <div>Email Tamu: ${escapeHtml(r.guest_email || "-")}</div>
          </div>
          <div>
            <strong>Detail Menginap:</strong>
            <div>Kamar: No. ${r.room ? escapeHtml(r.room.room_number) : r.room_id} (${r.room ? escapeHtml(r.room.type) : ""})</div>
            <div>Check-in: ${r.check_in_date.substring(0, 10)}</div>
            <div>Check-out: ${r.check_out_date.substring(0, 10)} (Pukul 12:00 WIB)</div>
            <div>Durasi: <strong>${r.nights} Malam</strong></div>
          </div>
        </div>

        <table class="table" style="margin-bottom: 1.5rem;">
          <thead>
            <tr>
              <th>Deskripsi</th>
              <th>Tarif / Malam</th>
              <th>Durasi</th>
              <th style="text-align: right;">Total</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>Sewa Kamar (${r.room ? escapeHtml(r.room.type) : "Hotel Room"})</td>
              <td>${formatIDR(r.room_rate)}</td>
              <td>${r.nights} Malam</td>
              <td style="text-align: right;"><strong>${formatIDR(r.total_amount)}</strong></td>
            </tr>
          </tbody>
          <tfoot>
            <tr>
              <td colspan="3" style="text-align: right;">Metode Pembayaran:</td>
              <td style="text-align: right; font-weight: 600;">${escapeHtml(paymentMethod)}</td>
            </tr>
            <tr>
              <td colspan="3" style="text-align: right;">Uang Muka (DP):</td>
              <td style="text-align: right; color: var(--secondary); font-weight: 600;">- ${formatIDR(r.down_payment)}</td>
            </tr>
            <tr>
              <td colspan="3" style="text-align: right; font-size: 1rem;"><strong>Sisa Tagihan:</strong></td>
              <td style="text-align: right; font-size: 1rem; color: var(--primary);"><strong>${formatIDR(r.remaining_amount)}</strong></td>
            </tr>
          </tfoot>
        </table>

        <div style="display: flex; justify-content: space-between; align-items: center; font-size: 0.85rem; color: var(--text-muted);">
          <div>
            Status Transaksi: <strong style="color: var(--primary);">${escapeHtml(r.status)}</strong> | 
            Status Bayar: <strong style="color: var(--secondary);">${escapeHtml(r.payment_status)}</strong> |
            Metode: <strong style="color: var(--text);">${escapeHtml(paymentMethod)}</strong>
          </div>
          <div style="text-align: right;">
            Terima kasih atas kunjungan Anda.
          </div>
        </div>
      </div>
    `;

    document.getElementById("invoiceModalBody").innerHTML = invoiceHTML;
    document.getElementById("invoiceModal").style.display = "flex";
  } catch (err) {
    alert("Gagal memuat invoice: " + err.message);
  }
}

function closeInvoiceModal() {
  document.getElementById("invoiceModal").style.display = "none";
}

function printInvoice() {
  window.print();
}
