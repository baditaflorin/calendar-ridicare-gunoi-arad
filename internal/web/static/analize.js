const staticMode = window.GUNOI_STATIC === true;
const siteRoot = new URL("../", document.currentScript.src);
const dataRoot = new URL("data/", siteRoot);

const wasteLabels = {
  residual: "Rezidual",
  plastic_metal: "Plastic & Metal",
  paper: "Hârtie & Carton",
  bio: "Bio",
  glass: "Sticlă",
  bulky: "Voluminoase",
  textile: "Textile",
  hazardous: "Periculoase"
};

const wasteColors = {
  residual: "#374151",
  plastic_metal: "#eab308",
  paper: "#2563eb",
  bio: "#16a34a",
  glass: "#10b981",
  bulky: "#7c3aed",
  textile: "#db2777",
  hazardous: "#dc2626"
};

document.addEventListener("DOMContentLoaded", async () => {
  try {
    const [catalog, stats] = await Promise.all([
      fetchJSON(new URL("catalog.json", dataRoot)),
      fetchJSON(new URL("stats.json", dataRoot))
    ]);
    renderSummaryCards(stats, catalog);
    renderNeighborhoods(catalog);
    renderWasteTypes(stats);
    renderTodayBreakdown(stats);
    renderWeeklyHeatmap(stats);
  } catch (err) {
    console.error("Failed to load stats:", err);
    document.querySelector(".stats-grid").innerHTML = `<p style="color: var(--muted);">Nu am putut incarca datele statistice. Incearca din nou.</p>`;
  }
});

async function fetchJSON(url) {
  const response = await fetch(url);
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return response.json();
}

function renderSummaryCards(stats, catalog) {
  const today = new Date().toISOString().slice(0, 10);
  const todayData = stats.daily_breakdown[today] || {};
  const todayTotal = Object.values(todayData).reduce((s, v) => s + v, 0);
  
  const summaryHTML = `
    <div class="stat-number-grid">
      <div class="stat-pill">
        <span class="stat-big">${stats.total_places.toLocaleString("ro")}</span>
        <span class="stat-label">Străzi monitorizate</span>
      </div>
      <div class="stat-pill">
        <span class="stat-big">${stats.total_events.toLocaleString("ro")}</span>
        <span class="stat-label">Evenimente programate</span>
      </div>
      <div class="stat-pill accent">
        <span class="stat-big">${todayTotal}</span>
        <span class="stat-label">Ridicări programate AZI</span>
      </div>
      <div class="stat-pill">
        <span class="stat-big">${(catalog.neighborhoods || []).length}</span>
        <span class="stat-label">Cartiere acoperite</span>
      </div>
    </div>
  `;
  document.getElementById("summary-cards").innerHTML = summaryHTML;
}

function renderNeighborhoods(catalog) {
  const ctx = document.getElementById("chart-neighborhoods").getContext("2d");
  const neighborhoods = catalog.neighborhoods || [];
  const sorted = [...neighborhoods].sort((a, b) => b.count - a.count).slice(0, 12);

  new Chart(ctx, {
    type: "bar",
    data: {
      labels: sorted.map(n => n.name),
      datasets: [{
        label: "Străzi",
        data: sorted.map(n => n.count),
        backgroundColor: "#2563eb",
        borderRadius: 6,
        borderSkipped: false
      }]
    },
    options: {
      indexAxis: "y",
      responsive: true,
      plugins: {
        legend: { display: false },
        tooltip: {
          callbacks: {
            label: ctx => `${ctx.raw} străzi`
          }
        }
      },
      scales: {
        x: { grid: { display: false } },
        y: { grid: { display: false } }
      }
    }
  });
}

function renderWasteTypes(stats) {
  const ctx = document.getElementById("chart-waste-types").getContext("2d");
  const counts = stats.waste_type_counts || {};
  const types = Object.keys(counts).sort((a, b) => counts[b] - counts[a]);

  new Chart(ctx, {
    type: "doughnut",
    data: {
      labels: types.map(t => wasteLabels[t] || t),
      datasets: [{
        data: types.map(t => counts[t]),
        backgroundColor: types.map(t => wasteColors[t] || "#94a3b8"),
        borderWidth: 2,
        borderColor: "#fff"
      }]
    },
    options: {
      responsive: true,
      cutout: "55%",
      plugins: {
        legend: { position: "bottom" },
        tooltip: {
          callbacks: {
            label: ctx => `${ctx.label}: ${ctx.raw.toLocaleString("ro")} evenimente`
          }
        }
      }
    }
  });
}

function renderTodayBreakdown(stats) {
  const today = new Date().toISOString().slice(0, 10);
  const todayData = stats.daily_breakdown[today] || {};
  const container = document.getElementById("today-detail");

  if (Object.keys(todayData).length === 0) {
    container.innerHTML = `
      <p style="font-size: 18px; text-align: center; padding: 20px; color: var(--muted);">
        Nicio colectare programată azi (${today}).
      </p>`;
    return;
  }

  const items = Object.entries(todayData)
    .sort((a, b) => b[1] - a[1])
    .map(([type, count]) => `
      <div class="today-row">
        <span class="today-dot" style="background: ${wasteColors[type] || '#94a3b8'}"></span>
        <span class="today-label">${wasteLabels[type] || type}</span>
        <span class="today-count">${count} ridicări</span>
      </div>
    `).join("");

  container.innerHTML = items;
}

function renderWeeklyHeatmap(stats) {
  const ctx = document.getElementById("chart-weekly").getContext("2d");
  const breakdown = stats.daily_breakdown || {};
  
  // Get next 14 days
  const days = [];
  const counts = [];
  const now = new Date();
  for (let i = 0; i < 14; i++) {
    const d = new Date(now);
    d.setDate(d.getDate() + i);
    const iso = d.toISOString().slice(0, 10);
    const dayData = breakdown[iso] || {};
    const total = Object.values(dayData).reduce((s, v) => s + v, 0);
    const shortDay = d.toLocaleDateString("ro", { weekday: "short", day: "numeric" });
    days.push(shortDay);
    counts.push(total);
  }

  new Chart(ctx, {
    type: "bar",
    data: {
      labels: days,
      datasets: [{
        label: "Ridicări programate",
        data: counts,
        backgroundColor: counts.map((c, i) => i === 0 ? "#2563eb" : "#93c5fd"),
        borderRadius: 4,
        borderSkipped: false
      }]
    },
    options: {
      responsive: true,
      plugins: {
        legend: { display: false },
        tooltip: {
          callbacks: {
            label: ctx => `${ctx.raw} ridicări`
          }
        }
      },
      scales: {
        x: { grid: { display: false } },
        y: { beginAtZero: true, grid: { color: "#f1f5f9" } }
      }
    }
  });
}
