// dashboard.js - Dashboard Metrik KPI Hotel & Rekapitulasi Laporan Keuangan (USALI)

async function loadDashboardMetrics() {
  try {
    const res = await authFetch("/api/reports/dashboard");
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || "Gagal memuat metrik dashboard");
    }

    const result = await res.json();
    const data = result.data;
    renderDashboard(data);
  } catch (err) {
    console.error("Gagal memuat dashboard:", err);
  }
}

function renderDashboard(data) {
  const formatIDR = (num) => new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(num);

  // 1. KPI Metrik Utama
  const kpiOccupancy = document.getElementById("kpiOccupancy");
  if (kpiOccupancy) kpiOccupancy.textContent = `${data.occupancy_rate || 0}%`;

  const kpiADR = document.getElementById("kpiADR");
  if (kpiADR) kpiADR.textContent = formatIDR(data.adr || 0);

  const kpiRevPAR = document.getElementById("kpiRevPAR");
  if (kpiRevPAR) kpiRevPAR.textContent = formatIDR(data.revpar || 0);

  // 2. Ringkasan Keuangan (Laba / Rugi)
  const metricRevenue = document.getElementById("metricRevenue");
  if (metricRevenue) metricRevenue.textContent = formatIDR(data.room_revenue || 0);

  const metricExpenses = document.getElementById("metricExpenses");
  if (metricExpenses) metricExpenses.textContent = formatIDR(data.total_expenses || 0);

  const metricNetIncome = document.getElementById("metricNetIncome");
  if (metricNetIncome) {
    const net = data.net_income || 0;
    metricNetIncome.textContent = formatIDR(net);
    metricNetIncome.style.color = net >= 0 ? "var(--success)" : "var(--danger)";
    
    const badgeNet = document.getElementById("badgeNetIncomeStatus");
    if (badgeNet) {
      badgeNet.textContent = net >= 0 ? "LABA BERSIH" : "RUGI BERSIH";
      badgeNet.className = net >= 0 ? "badge badge-success" : "badge badge-danger";
    }
  }

  // 3. Status Inventaris Kamar
  const invTotal = document.getElementById("invTotalRooms");
  if (invTotal) invTotal.textContent = data.total_rooms || 0;

  const invAvailable = document.getElementById("invAvailableRooms");
  if (invAvailable) invAvailable.textContent = data.available_rooms || 0;

  const invOccupied = document.getElementById("invOccupiedRooms");
  if (invOccupied) invOccupied.textContent = data.occupied_rooms || 0;

  const invCleaning = document.getElementById("invCleaningRooms");
  if (invCleaning) invCleaning.textContent = data.cleaning_rooms || 0;

  const invMaintenance = document.getElementById("invMaintenanceRooms");
  if (invMaintenance) invMaintenance.textContent = data.maintenance_rooms || 0;

  // 4. Breakdown Pengeluaran USALI
  const tbody = document.getElementById("usaliBreakdownTableBody");
  if (tbody) {
    const items = data.expenses_by_category || [];
    if (items.length === 0) {
      tbody.innerHTML = `<tr><td colspan="3" style="text-align: center; color: var(--text-muted); padding: 1.5rem;">Belum ada data pengeluaran operasional</td></tr>`;
    } else {
      tbody.innerHTML = items.map(cat => {
        const pct = data.total_expenses > 0 ? ((cat.total / data.total_expenses) * 100).toFixed(1) : 0;
        return `
          <tr>
            <td><strong>${escapeHtml(cat.category)}</strong></td>
            <td><strong style="color: var(--danger);">${formatIDR(cat.total)}</strong></td>
            <td>
              <div style="display: flex; align-items: center; gap: 0.5rem;">
                <div style="flex: 1; background: #e2e8f0; height: 8px; border-radius: 4px; overflow: hidden;">
                  <div style="background: var(--danger); width: ${pct}%; height: 100%;"></div>
                </div>
                <small style="color: var(--text-muted); width: 45px;">${pct}%</small>
              </div>
            </td>
          </tr>
        `;
      }).join("");
    }
  }
}
