const printMonthNames = ["Ianuarie", "Februarie", "Martie", "Aprilie", "Mai", "Iunie", "Iulie", "August", "Septembrie", "Octombrie", "Noiembrie", "Decembrie"];
const printWeekdays = ["Luni", "Marti", "Miercuri", "Joi", "Vineri", "Sambata", "Duminica"];
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
  const sources = unique(events.map((event) => event.source_url).filter(Boolean));
  const updated = events.find((event) => event.fetched_at)?.fetched_at;
  document.querySelector("#print-paper").innerHTML = `
    <header class="paper-head">
      <div class="brand print-brand">
        <span class="brand-mark" aria-hidden="true"></span>
        <span>Gunoi <strong>Arad</strong></span>
      </div>
      <span>Programul tau. Orasul nostru.</span>
    </header>
    <h1 style="letter-spacing: -0.5px; word-spacing: 2px;">Calendar de colectare &mdash; Strada ${escapeHTML(place.street_raw)}, ${escapeHTML(place.cartier || "Arad")}</h1>
    <h2 style="margin-bottom: 18px;">${printMonthNames[month.getMonth()]} ${month.getFullYear()}</h2>
    <div class="print-calendar">
      ${printWeekdays.map((day) => `<div class="print-weekday">${day}</div>`).join("")}
      ${days.map((day) => `
        <div class="print-day ${day.inMonth ? "" : "muted"}">
          <span>${day.date.getDate()}</span>
          ${day.events.map((event) => `<div class="print-event"><i style="background: ${event.color}"></i>${escapeHTML(event.label)}${event.confidence < 1 ? " *" : ""}</div>`).join("")}
        </div>
      `).join("")}
    </div>
    <div class="print-legend">
      ${legend(events).map((event) => `<span><i style="background: ${event.color}"></i>${escapeHTML(event.label)}</span>`).join("")}
    </div>
    <section class="print-notes">
      <div>
        <h3>Instructiuni rapide</h3>
        <p>Scoate recipientele pana la ora 07:00. Respecta tipul de deseu pentru colectare separata.</p>
      </div>
      <div>
        <h3>Trasabilitate</h3>
        <p>Datele au link catre sursa oficiala. Evenimentele marcate cu * sunt inferate din programul cartierului.</p>
      </div>
      <div>
        <h3>Actualizare</h3>
        <p>${updated ? `Sursa verificata pe ${formatDate(updated)}.` : "Sursa verificata automat."}</p>
        ${sources.slice(0, 2).map((url) => `<p><a href="${url}" target="_blank" rel="noreferrer">Verifica sursa oficiala</a></p>`).join("")}
      </div>
    </section>
    <footer class="print-credit">
      <div style="display: flex; justify-content: space-between; align-items: flex-end;">
        <div>
          <p>Date publice RETIM si Primaria Arad. Calendar independent pentru uz civic.</p>
          <p>Dezvoltat pro-bono de <strong>Florin Badita</strong> (baditaflorin@gmail.com)</p>
        </div>
        <div style="text-align: right;">
          <p>Sustine proiectul: <strong>paypal.me/florinbadita</strong></p>
          <p>baditaflorin.github.io/calendar-ridicare-gunoi-arad/</p>
        </div>
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
    paper: "Hartie & Carton",
    plastic_metal: "Plastic & Metal",
    glass: "Sticla",
    bulky: "Voluminoase",
    textile: "Textile",
    hazardous: "Periculoase",
  }[type] || type;
}

function wasteColor(type) {
  return {
    residual: "#374151",
    bio: "#16a34a",
    paper: "#2563eb",
    plastic_metal: "#eab308",
    glass: "#16a34a",
    bulky: "#7c3aed",
    textile: "#db2777",
    hazardous: "#dc2626",
  }[type] || "#64748b";
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
