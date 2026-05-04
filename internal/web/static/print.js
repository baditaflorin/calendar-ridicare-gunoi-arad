const printMonthNames = ["Ianuarie", "Februarie", "Martie", "Aprilie", "Mai", "Iunie", "Iulie", "August", "Septembrie", "Octombrie", "Noiembrie", "Decembrie"];
const printWeekdays = ["Luni", "Marți", "Miercuri", "Joi", "Vineri", "Sâmbătă", "Duminică"];
const printSiteRoot = new URL("../", document.currentScript.src);
const printDataRoot = new URL("data/", printSiteRoot);

document.addEventListener("DOMContentLoaded", async () => {
  const params = new URLSearchParams(window.location.search);
  const placeId = params.get("place_id");
  const monthParam = params.get("month") || new Date().toISOString().slice(0, 7);
  if (!placeId) {
    document.querySelector("#print-paper").innerHTML = `<p>Nu am primit strada pentru calendar.</p>`;
    return;
  }
  const payload = await fetchJSON(new URL(`places/${placeId}.json`, printDataRoot));
  const month = parseMonth(monthParam);
  const events = payload.events.map(decorateEvent).filter((event) => event.date.startsWith(monthParam));
  document.querySelector("#print-back").href = new URL(`program/?place_id=${placeId}`, printSiteRoot).toString();
  renderPrint(payload.place, events, month);
});

function renderPrint(place, events, month) {
  const days = buildMonthGrid(month, events);

  const wasteEmoji = {
    residual: "🗑️", bio: "🌿", paper: "📄",
    plastic_metal: "♻️", glass: "🍃", bulky: "📦",
    textile: "👕", hazardous: "⚠️"
  };

  const wasteBg = {
    residual: "#f1f5f9", bio: "#d1fae5", paper: "#dbeafe",
    plastic_metal: "#fef3c7", glass: "#d1fae5", bulky: "#ede9fe",
    textile: "#fce7f3", hazardous: "#fee2e2"
  };

  const recycleLogoSVG = `<svg width="36" height="36" viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="32" cy="32" r="32" fill="#d1fae5"/>
    <path d="M32 14 L38 24 H26 Z" fill="#10b981"/>
    <path d="M44 42 L38 32 L50 32 Z" fill="#059669"/>
    <path d="M20 42 L26 32 L14 32 Z" fill="#34d399"/>
    <circle cx="32" cy="32" r="6" fill="#fff"/>
    <circle cx="32" cy="32" r="4" fill="#10b981"/>
  </svg>`;

  document.querySelector("#print-paper").innerHTML = `
    <header class="paper-head">
      <div class="brand print-brand">
        ${recycleLogoSVG}
        <div class="brand-text">
          <span class="brand-name">Reciclare <strong>Arad</strong></span>
          <span class="brand-tagline">🌍 Împreună salvăm planeta!</span>
        </div>
      </div>
      <div class="print-address-block">
        <div class="print-street">${escapeHTML(place.street_raw)}</div>
        <div class="print-district">${escapeHTML(place.cartier || "Arad")}</div>
        <div class="print-month-label">${printMonthNames[month.getMonth()]} ${month.getFullYear()}</div>
      </div>
    </header>

    <div class="print-calendar">
      ${printWeekdays.map((day) => `<div class="print-weekday">${day}</div>`).join("")}
      ${days.map((day) => `
        <div class="print-day ${day.inMonth ? "" : "muted"}">
          <span class="day-num">${day.date.getDate()}</span>
          <div class="day-events">
            ${day.events.map((event) => `
              <div class="print-event" style="background:${wasteBg[event.waste_type] || "#f1f5f9"}; border-left: 3px solid ${event.color}">
                <span class="event-emoji">${wasteEmoji[event.waste_type] || "♻️"}</span>
                <span class="event-label">${escapeHTML(event.label)}</span>
              </div>
            `).join("")}
          </div>
        </div>
      `).join("")}
    </div>

    <footer class="print-footer cute-footer">
      <div class="footer-left">
        <p><strong>Dezvoltat pro-bono de Florin Badita</strong> (baditaflorin@gmail.com) &bull; paypal.me/florinbadita</p>
        <p>baditaflorin.github.io/calendar-ridicare-gunoi-arad/ &mdash; Date: RETIM &amp; Primăria Arad</p>
      </div>
      <div class="footer-legend">
        <span>🌿 Bio</span><span>📄 Hârtie</span><span>♻️ Plastic</span><span>🗑️ Rezidual</span>
      </div>
    </footer>
  `;
}

function decorateEvent(event) {
  return {
    ...event,
    label: wasteLabel(event.waste_type),
    color: wasteColor(event.waste_type),
  };
}

function wasteLabel(type) {
  return {
    residual: "Rezidual",
    bio: "Bio",
    paper: "Hârtie & Carton",
    plastic_metal: "Plastic & Metal",
    glass: "Sticlă",
    bulky: "Voluminoase",
    textile: "Textile",
    hazardous: "Periculoase",
  }[type] || type;
}

function wasteColor(type) {
  return {
    residual: "#475569", // Darker slate
    bio: "#10b981",     // Emerald
    paper: "#3b82f6",     // Blue
    plastic_metal: "#f59e0b", // Amber
    glass: "#059669",    // Green
    bulky: "#8b5cf6",    // Violet
    textile: "#ec4899",   // Pink
    hazardous: "#ef4444",  // Red
  }[type] || "#94a3b8";
}

function buildMonthGrid(month, events) {
  const first = new Date(month.getFullYear(), month.getMonth(), 1);
  const offset = (first.getDay() + 6) % 7;
  const start = addDays(first, -offset);
  const byDate = new Map();
  for (const event of events) {
    if (!byDate.has(event.date)) byDate.set(event.date, []);
    byDate.get(event.date).push(event);
  }
  return Array.from({ length: 42 }, (_, index) => {
    const date = addDays(start, index);
    return {
      date,
      inMonth: date.getMonth() === month.getMonth(),
      events: byDate.get(isoDate(date)) || [],
    };
  });
}

function legend(events) {
  const seen = new Set();
  return events.filter((event) => {
    if (seen.has(event.waste_type)) return false;
    seen.add(event.waste_type);
    return true;
  });
}

async function fetchJSON(url) {
  const response = await fetch(url);
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return response.json();
}

function parseMonth(value) {
  const [year, month] = value.split("-").map(Number);
  return new Date(year, month - 1, 1);
}

function addDays(date, days) {
  const next = new Date(date);
  next.setDate(next.getDate() + days);
  return next;
}

function isoDate(date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

function formatDate(value) {
  const date = new Date(value);
  return `${date.getDate()} ${printMonthNames[date.getMonth()]} ${date.getFullYear()}`;
}

function unique(values) {
  return Array.from(new Set(values));
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

window.downloadImage = async function() {
  const paper = document.querySelector("#print-paper");
  const btn = document.querySelector(".print-toolbar button");
  const originalText = btn.textContent;
  
  if (!window.html2canvas) {
    alert("Libraria de generare imagine nu s-a incarcat inca. Te rugam sa mai astepti o secunda.");
    return;
  }

  try {
    btn.textContent = "Se genereaza imaginea...";
    btn.disabled = true;
    
    const canvas = await html2canvas(paper, {
      scale: 2,
      useCORS: true,
      allowTaint: true,
      backgroundColor: "#ffffff",
      windowWidth: 1200,
      logging: true,
    });
    
    const image = canvas.toDataURL("image/png", 1.0);
    const link = document.createElement("a");
    
    // try to get street name for filename
    const heading = paper.querySelector("h1")?.textContent || "";
    const streetMatch = heading.match(/Strada (.*)/);
    const filename = streetMatch ? `calendar-${streetMatch[1].replace(/[^a-z0-9]/gi, '_').toLowerCase()}.png` : `calendar-gunoi-arad.png`;
    
    link.download = filename;
    link.href = image;
    link.click();
  } catch(err) {
    console.error("Failed to generate image", err);
    alert("Eroare la generarea imaginii: " + err.message);
  } finally {
    btn.textContent = originalText;
    btn.disabled = false;
  }
};
