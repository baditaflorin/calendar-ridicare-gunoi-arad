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

const mascots = {
  residual: "🛰️",
  bio: "🍀",
  paper: "📘",
  plastic_metal: "⚡",
  glass: "✨",
  bulky: "🦖",
  textile: "👕",
  hazardous: "⚠️"
};

const mascotNames = {
  residual: "Shadow",
  bio: "Leafy",
  paper: "Bluey",
  plastic_metal: "Sparky",
  glass: "Glaze",
  bulky: "Rex",
  textile: "Cotton",
  hazardous: "Danger"
};

function renderPrint(place, events, month) {
  const days = buildMonthGrid(month, events);
  const sources = unique(events.map((event) => event.source_url).filter(Boolean));
  const updated = events.find((event) => event.fetched_at)?.fetched_at;
  
  document.querySelector("#print-paper").innerHTML = `
    <div class="sky-accent"></div>
    <header class="paper-head">
      <div class="brand print-brand">
        <span class="brand-mark" aria-hidden="true"></span>
        <span>Reciclare <strong>Arad</strong></span>
      </div>
      <div class="header-motto">
        <span>Aventura Reciclării</span>
        <span class="motto-sub">Împreună salvăm planeta! 🌍</span>
      </div>
    </header>

    <div class="print-meta">
      <h1>Strada ${escapeHTML(place.street_raw)} &bull; <span>${escapeHTML(place.cartier || "Arad")}</span></h1>
      <h2 class="cute-month">${printMonthNames[month.getMonth()]} ${month.getFullYear()}</h2>
    </div>

    <div class="print-calendar">
      ${printWeekdays.map((day) => `<div class="print-weekday">${day}</div>`).join("")}
      ${days.map((day) => `
        <div class="print-day ${day.inMonth ? "" : "muted"}">
          <span class="day-num">${day.date.getDate()}</span>
          <div class="day-events">
            ${day.events.map((event) => `
              <div class="print-event cute-event" style="border-left: 4px solid ${event.color}">
                <span class="event-mascot">${mascots[event.waste_type] || "♻️"}</span>
                <div class="event-info">
                  <span class="event-label">${escapeHTML(event.label)}</span>
                  <span class="event-check">⭐ Am ajutat!</span>
                </div>
              </div>
            `).join("")}
          </div>
        </div>
      `).join("")}
    </div>

    <div class="print-bottom-row">
      <div class="print-legend-cute">
        <h3>Echipa Reciclării</h3>
        <div class="legend-items">
          ${legend(events).map((event) => `
            <div class="legend-item">
              <span class="l-mascot">${mascots[event.waste_type] || "♻️"}</span>
              <div class="l-text">
                <span class="l-name">${mascotNames[event.waste_type] || "Erou"}</span>
                <span class="l-label">${escapeHTML(event.label)}</span>
              </div>
            </div>
          `).join("")}
        </div>
      </div>

      <div class="color-me-box">
        <div class="color-me-art">
          <svg viewBox="0 0 100 100" width="80" height="80">
            <circle cx="50" cy="50" r="45" fill="none" stroke="#e2e8f0" stroke-width="2" />
            <path d="M30 40 Q50 20 70 40" fill="none" stroke="#cbd5e1" stroke-width="2" />
            <circle cx="40" cy="45" r="3" fill="#cbd5e1" />
            <circle cx="60" cy="45" r="3" fill="#cbd5e1" />
            <path d="M40 65 Q50 75 60 65" fill="none" stroke="#cbd5e1" stroke-width="2" />
          </svg>
        </div>
        <p>Colorează-mă!</p>
      </div>
    </div>

    <div class="mission-banner">
      <span class="mission-icon">🏆</span>
      <div class="mission-text">
        <strong>Misiunea Lunii:</strong> Colectează separat și bifează toate stelele de mai sus pentru a deveni un Erou Verde!
      </div>
    </div>

    <footer class="print-footer cute-footer">
      <div class="footer-left">
        <p>Creat cu ❤️ pentru copiii din Arad.</p>
        <p>baditaflorin.github.io/calendar-ridicare-gunoi-arad/</p>
      </div>
      <div class="footer-right">
        <p>Date RETIM & Primăria Arad</p>
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
    residual: "Rezidual (Shadow)",
    bio: "Bio (Leafy)",
    paper: "Hârtie (Bluey)",
    plastic_metal: "Plastic (Sparky)",
    glass: "Sticlă (Glaze)",
    bulky: "Voluminoase (Rex)",
    textile: "Textile (Cotton)",
    hazardous: "Periculoase (Danger)",
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
