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

  // Big friendly emoji for each type
  const wasteEmoji = {
    residual: "🗑️", bio: "🌱", paper: "📰",
    plastic_metal: "🧴", glass: "🫙", bulky: "📦",
    textile: "🎀", hazardous: "⚠️"
  };

  // Vibrant pastel backgrounds
  const wasteBg = {
    residual: "#e2e8f0", bio: "#bbf7d0", paper: "#bfdbfe",
    plastic_metal: "#fde68a", glass: "#a7f3d0", bulky: "#ddd6fe",
    textile: "#fbcfe8", hazardous: "#fecaca"
  };

  const wasteTextColor = {
    residual: "#334155", bio: "#065f46", paper: "#1e40af",
    plastic_metal: "#92400e", glass: "#065f46", bulky: "#5b21b6",
    textile: "#9d174d", hazardous: "#991b1b"
  };

  // Rainbow colors per weekday column (Mon–Sun)
  const dayColors = ["#fde68a","#bbf7d0","#bfdbfe","#ddd6fe","#fbcfe8","#fed7aa","#d1fae5"];
  const dayTextColors = ["#78350f","#065f46","#1e40af","#5b21b6","#9d174d","#7c2d12","#065f46"];

  const logoSVG = `<svg width="52" height="52" viewBox="0 0 80 80" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="40" cy="40" r="40" fill="#d1fae5"/>
    <!-- Leaf left -->
    <ellipse cx="22" cy="44" rx="12" ry="7" fill="#34d399" transform="rotate(-30 22 44)"/>
    <!-- Leaf right -->
    <ellipse cx="58" cy="44" rx="12" ry="7" fill="#10b981" transform="rotate(30 58 44)"/>
    <!-- Center stem -->
    <rect x="37" y="28" width="6" height="22" rx="3" fill="#059669"/>
    <!-- Smile -->
    <path d="M28 54 Q40 64 52 54" stroke="#059669" stroke-width="3" fill="none" stroke-linecap="round"/>
    <!-- Eyes -->
    <circle cx="32" cy="46" r="3" fill="#059669"/>
    <circle cx="48" cy="46" r="3" fill="#059669"/>
    <!-- Stars decorations -->
    <text x="6" y="22" font-size="12">⭐</text>
    <text x="58" y="18" font-size="10">✨</text>
  </svg>`;

  const decorStars = `<div class="deco-stars" aria-hidden="true">
    <span style="top:2mm;left:8mm;font-size:16px;color:#fbbf24;">⭐</span>
    <span style="top:6mm;left:22mm;font-size:10px;color:#34d399;">🌿</span>
    <span style="top:1mm;right:30mm;font-size:12px;color:#f472b6;">✨</span>
    <span style="top:5mm;right:14mm;font-size:10px;color:#60a5fa;">🌍</span>
  </div>`;

  document.querySelector("#print-paper").innerHTML = `
    <header class="paper-head">
      ${decorStars}
      <div class="brand print-brand">
        ${logoSVG}
        <div class="brand-text">
          <span class="brand-name">Reciclare <strong>Arad</strong></span>
          <span class="brand-tagline">🌱 Eroi ai Planetei!</span>
        </div>
      </div>
      <div class="print-address-block">
        <div class="print-street">📍 ${escapeHTML(place.street_raw)}</div>
        <div class="print-district">🏘️ ${escapeHTML(place.cartier || "Arad")}</div>
        <div class="print-month-label">🗓️ ${printMonthNames[month.getMonth()]} ${month.getFullYear()}</div>
      </div>
    </header>

    <div class="print-calendar">
      ${printWeekdays.map((day) => `
        <div class="print-weekday">${day}</div>
      `).join("")}
      ${days.map((day) => {
        const isWeekend = day.date.getDay() === 0 || day.date.getDay() === 6;
        return `
        <div class="print-day ${day.inMonth ? "" : "muted"}${isWeekend ? " weekend" : ""}">
          <span class="day-num">${day.date.getDate()}</span>
          <div class="day-events">
            ${day.events.map((event) => `
              <div class="print-event" style="background:${wasteBg[event.waste_type] || "#e2e8f0"};color:${wasteTextColor[event.waste_type] || "#334155"}">
                <span class="event-emoji">${wasteEmoji[event.waste_type] || "♻️"}</span>
                <span class="event-label">${escapeHTML(event.label)}</span>
              </div>
            `).join("")}
          </div>
        </div>`;
      }).join("")}
    </div>

    <footer class="print-footer cute-footer">
      <div class="footer-left">
        🌟 <strong>Florin Badita</strong> · baditaflorin@gmail.com · paypal.me/florinbadita · baditaflorin.github.io/calendar-ridicare-gunoi-arad/
      </div>
      <div class="footer-legend">
        <span>🌱 Bio</span><span>📰 Hârtie</span><span>🧴 Plastic</span><span>🗑️ Rezidual</span><span>🫙 Sticlă</span>
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
