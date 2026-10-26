// expenses.js - Manajemen Biaya Operasional Berstandar USALI

let currentExpenses = [];
let usaliCategories = [];

// Muat daftar kategori USALI dari server
async function loadExpenseCategories() {
  try {
    const res = await authFetch("/api/expenses/categories");
    const data = await res.json();
    usaliCategories = data.data || [];

    // Isi dropdown kategori pada modal form
    const select = document.getElementById("expCategory");
    if (select) {
      select.innerHTML = usaliCategories.map(cat => 
        `<option value="${escapeHtml(cat)}">${escapeHtml(cat)}</option>`
      ).join("");
    }

    // Isi tombol filter kategori
    const filterContainer = document.getElementById("expenseFilterBar");
    if (filterContainer) {
      filterContainer.innerHTML = `
        <button class="btn btn-outline btn-sm active-filter" onclick="filterExpenses('')">Semua Kategori</button>
        ${usaliCategories.map(cat => 
          `<button class="btn btn-outline btn-sm" onclick="filterExpenses('${escapeHtml(cat)}')">${escapeHtml(cat)}</button>`
        ).join("")}
      `;
    }
  } catch (err) {
    console.error("Gagal memuat kategori USALI:", err);
  }
}

// Muat data pengeluaran
async function loadExpenses(categoryFilter = "") {
  try {
    let url = "/api/expenses";
    if (categoryFilter) {
      url += `?category=${encodeURIComponent(categoryFilter)}`;
    }

    const res = await authFetch(url);
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || "Gagal memuat pengeluaran");
    }

    const result = await res.json();
    currentExpenses = result.data || [];
    renderExpenses(currentExpenses, result.total_amount || 0);
  } catch (err) {
    console.error("Gagal memuat biaya operasional:", err);
    const tbody = document.getElementById("expensesTableBody");
    if (tbody) {
      tbody.innerHTML = `<tr><td colspan="6" style="text-align: center; color: var(--danger); padding: 1.5rem;">${err.message}</td></tr>`;
    }
  }
}

function filterExpenses(category) {
  loadExpenses(category);
}

// Render data pengeluaran
function renderExpenses(expenses, totalAmount) {
  const tbody = document.getElementById("expensesTableBody");
  const totalDisplay = document.getElementById("totalExpenseDisplay");
  if (totalDisplay) {
    const formatIDR = (num) => new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(num);
    totalDisplay.textContent = formatIDR(totalAmount);
  }

  if (!tbody) return;

  if (expenses.length === 0) {
    tbody.innerHTML = `
      <tr>
        <td colspan="6" style="text-align: center; color: var(--text-muted); padding: 2rem;">
          Belum ada catatan biaya operasional. Klik tombol "Catat Pengeluaran" untuk menambahkan.
        </td>
      </tr>
    `;
    return;
  }

  const formatIDR = (num) => new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(num);

  tbody.innerHTML = expenses.map(e => {
    const expDate = e.expense_date.substring(0, 10);
    return `
      <tr>
        <td><strong>${expDate}</strong></td>
        <td><span class="badge badge-secondary">${escapeHtml(e.category)}</span></td>
        <td>
          <div><strong>${escapeHtml(e.title)}</strong></div>
          <small style="color: var(--text-muted);">${escapeHtml(e.receipt_no ? 'Bukti/Kuitansi: ' + e.receipt_no : '-')}</small>
        </td>
        <td><strong style="color: var(--danger);">${formatIDR(e.amount)}</strong></td>
        <td><small style="color: var(--text-muted);">${escapeHtml(e.notes || "-")}</small></td>
        <td>
          <button class="btn-sm btn-outline" onclick="openEditExpenseModal(${e.id})">Edit</button>
          <button class="btn-sm btn-danger" onclick="deleteExpense(${e.id}, '${escapeHtml(e.title)}')">Hapus</button>
        </td>
      </tr>
    `;
  }).join("");
}

// Buka modal tambah pengeluaran baru
function openAddExpenseModal() {
  document.getElementById("expenseModalTitle").textContent = "Catat Pengeluaran Operasional (USALI)";
  document.getElementById("expenseId").value = "";
  document.getElementById("expTitle").value = "";
  document.getElementById("expAmount").value = "";
  document.getElementById("expReceiptNo").value = "";
  document.getElementById("expNotes").value = "";
  document.getElementById("expDate").value = new Date().toISOString().split("T")[0];
  document.getElementById("expenseModalAlert").style.display = "none";
  document.getElementById("expenseModal").style.display = "flex";
}

// Buka modal edit pengeluaran
function openEditExpenseModal(id) {
  const expense = currentExpenses.find(e => e.id === id);
  if (!expense) return;

  document.getElementById("expenseModalTitle").textContent = "Edit Biaya Operasional";
  document.getElementById("expenseId").value = expense.id;
  document.getElementById("expTitle").value = expense.title;
  document.getElementById("expCategory").value = expense.category;
  document.getElementById("expAmount").value = expense.amount;
  document.getElementById("expDate").value = expense.expense_date.substring(0, 10);
  document.getElementById("expReceiptNo").value = expense.receipt_no || "";
  document.getElementById("expNotes").value = expense.notes || "";
  document.getElementById("expenseModalAlert").style.display = "none";
  document.getElementById("expenseModal").style.display = "flex";
}

function closeExpenseModal() {
  document.getElementById("expenseModal").style.display = "none";
}

// Simpan pengeluaran (POST atau PUT)
async function handleExpenseFormSubmit(e) {
  e.preventDefault();
  const alertEl = document.getElementById("expenseModalAlert");
  alertEl.style.display = "none";

  const id = document.getElementById("expenseId").value;
  const payload = {
    category: document.getElementById("expCategory").value,
    title: document.getElementById("expTitle").value.trim(),
    amount: parseFloat(document.getElementById("expAmount").value),
    expense_date: document.getElementById("expDate").value,
    receipt_no: document.getElementById("expReceiptNo").value.trim(),
    notes: document.getElementById("expNotes").value.trim()
  };

  try {
    const isEdit = !!id;
    const url = isEdit ? `/api/expenses/${id}` : "/api/expenses";
    const method = isEdit ? "PUT" : "POST";

    const res = await authFetch(url, {
      method: method,
      body: JSON.stringify(payload)
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || "Gagal menyimpan pengeluaran");
    }

    closeExpenseModal();
    loadExpenses();
  } catch (err) {
    alertEl.textContent = err.message;
    alertEl.style.display = "block";
  }
}

// Hapus pengeluaran
async function deleteExpense(id, title) {
  if (!confirm(`Hapus catatan biaya operasional "${title}"?`)) {
    return;
  }

  try {
    const res = await authFetch(`/api/expenses/${id}`, {
      method: "DELETE"
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || "Gagal menghapus pengeluaran");
    }

    loadExpenses();
  } catch (err) {
    alert("Error: " + err.message);
  }
}
