const state = {
  place: null,
  events: [],
  month: new Date(),
  neighborhoods: [],
  catalog: null,
};

const staticMode = window.GUNOI_STATIC === true;
const siteRoot = new URL("../", document.currentScript.src);
const dataRoot = new URL("data/", siteRoot);

const wasteGlyph = {
  paper: "□",
  plastic_metal: "♙",
  bio: "◜",
  residual: "▱",
  bulky: "◆",
  textile: "◇",
  hazardous: "!",
};

const weekdays = ["L", "MA", "MI", "J", "V", "S", "D"];
const fullWeekdays = ["Luni", "Marti", "Miercuri", "Joi", "Vineri", "Sambata", "Duminica"];
const monthNames = ["Ianuarie", "Februarie", "Martie", "Aprilie", "Mai", "Iunie", "Iulie", "August", "Septembrie", "Octombrie", "Noiembrie", "Decembrie"];

document.addEventListener("DOMContentLoaded", () => {
  const shell = document.querySelector(".shell");
  window.addEventListener("scroll", () => {
    const scrolled = window.pageYOffset;
    if (shell) {
      shell.style.setProperty("--parallax-offset", `${scrolled * 0.4}px`);
    }
  });
  const params = new URLSearchParams(window.location.search);
  const placeId = params.get("place_id");
  bindSearch();
  bindButtons();
  loadNeighborhoods();
  if (placeId) {
    loadPlace(placeId);
  }
});

function bindSearch() {
  const form = document.querySelector("#search-form");
  const input = document.querySelector("#street-search");
  let timer;
  form.addEventListener("submit", (event) => {
    event.preventDefault();
    search(input.value, true);
  });
  input.addEventListener("input", () => {
    clearTimeout(timer);
    timer = setTimeout(() => search(input.value, false), 160);
  });
  document.querySelector("#neighborhood-select").addEventListener("change", () => {
    search(input.value, false);
  });
}

function bindButtons() {
  document.querySelector("#prev-month").addEventListener("click", () => {
    state.month = new Date(state.month.getFullYear(), state.month.getMonth() - 1, 1);
    loadEventsForCurrentWindow();
  });
  document.querySelector("#next-month").addEventListener("click", () => {
    state.month = new Date(state.month.getFullYear(), state.month.getMonth() + 1, 1);
    loadEventsForCurrentWindow();
  });
  document.querySelector("#share-button").addEventListener("click", async () => {
    if (!state.place) return;
    const url = programURL(state.place.id);
    if (navigator.share) {
      await navigator.share({ title: "Gunoi Arad", url });
    } else if (navigator.clipboard) {
      await navigator.clipboard.writeText(url);
    }
  });
}

async function search(query, pickFirst) {
  const panel = document.querySelector("#search-results");
  const neighborhood = document.querySelector("#neighborhood-select").value;
  if (!query.trim() && !neighborhood) {
    panel.hidden = true;
    return;
  }
  const results = await searchData(query, neighborhood, 12);
  if (pickFirst && results[0]) {
    panel.hidden = true;
    loadPlace(results[0].place.id);
    return;
  }
  panel.innerHTML = results.map((result) => `
    <button type="button" data-id="${result.place.id}">
      <span>${escapeHTML(result.place.street_raw)}, ${escapeHTML(result.place.cartier || "Arad")}</span>
      <small>${result.score}</small>
    </button>
  `).join("");
  panel.hidden = results.length === 0;
  panel.querySelectorAll("button").forEach((button) => {
    button.addEventListener("click", () => {
      panel.hidden = true;
      loadPlace(button.dataset.id);
    });
  });
}

async function loadNeighborhoods() {
  state.neighborhoods = await neighborhoodsData();
  const select = document.querySelector("#neighborhood-select");
  select.insertAdjacentHTML("beforeend", state.neighborhoods.map((item) => (
    `<option value="${escapeHTML(item.norm)}">${escapeHTML(item.name)} (${item.count})</option>`
  )).join(""));
  const total = state.neighborhoods.reduce((sum, item) => sum + item.count, 0);
  document.querySelector("#catalog-count").textContent = `${total} strazi reale din ${state.neighborhoods.length} cartiere`;
}

async function loadPlace(placeId) {
  const data = await placeData(placeId);
  state.place = data.place;
  state.events = data.events || [];
  document.querySelector("#street-search").value = `${state.place.street_raw}, ${state.place.cartier || "Arad"}`;
  window.history.replaceState(null, "", programURL(state.place.id));
  renderAll();
}

function loadEventsForCurrentWindow() {
  if (state.place) loadPlace(state.place.id);
}

function renderAll() {
  const place = state.place;
  // Reveal the program section
  const programSection = document.querySelector("#program");
  programSection.hidden = false;
  programSection.style.animation = "fadeSlideIn 0.5s ease-out";
  // Shrink the hero
  document.querySelector("#hero").classList.add("hero-compact");

  document.querySelector("#place-title").textContent = `Programul pentru Strada ${place.street_raw}, Arad`;
  document.querySelector("#place-subtitle").textContent = place.cartier ? `Cartier ${place.cartier}` : "Municipiul Arad";
  document.querySelector("#freshness").textContent = "actualizat";
  document.querySelector("#print-link").href = printURL();
  const icsLink = document.querySelector("#ics-link");
  if (staticMode) {
    icsLink.href = "#";
    icsLink.onclick = (event) => {
      event.preventDefault();
      downloadICS();
    };
  } else {
    icsLink.href = `/ics?place_id=${place.id}`;
    icsLink.onclick = null;
  }
  const official = state.events.find((event) => event.source_url);
  if (official) {
    document.querySelector("#official-link").href = official.source_url;
  }
  renderNext();
  renderLegend();
  renderCalendar();
  renderSources();
  // Scroll to the program section smoothly
  programSection.scrollIntoView({ behavior: "smooth", block: "start" });
}

function renderNext() {
  const next = upcomingEvents()[0];
  const card = document.querySelector("#next-card");
  if (!next) {
    card.innerHTML = `<div><p>Nu exista ridicari in intervalul incarcat.</p></div>`;
    return;
  }
  const date = parseDate(next.date);
  card.className = "next-card"; // Reset just in case
  const imgURL = staticMode ? new URL(`static/images/${wasteImage(next.waste_type)}`, siteRoot) : `/static/images/${wasteImage(next.waste_type)}`;
  
  card.innerHTML = `
    <div class="next-content">
      <p class="eyebrow">URMATOAREA COLECTARE</p>
      <h3>${relativeDate(date)}</h3>
      <p class="next-type" style="color: ${next.color};"><strong>${escapeHTML(next.label)}</strong></p>
      <p class="next-time">Incepand cu ${next.start_time || "07:00"}</p>
    </div>
    <div class="bin-visual" aria-hidden="true">
      <img src="${imgURL}" alt="Pubela ${escapeHTML(next.label)}" loading="lazy" />
    </div>
  `;
}

function renderWeek() {
  const strip = document.querySelector("#week-strip");
  const start = weekStart(new Date());
  strip.innerHTML = "";
  for (let i = 0; i < 7; i += 1) {
    const day = addDays(start, i);
    const events = eventsForDate(day);
    strip.insertAdjacentHTML("beforeend", `
      <div class="week-day">
        <span class="week-label">${weekdays[i]}</span>
        <span class="week-number">${day.getDate()}</span>
        <span class="dots">${events.map((event) => `<i class="dot" style="background:${event.color}"></i>`).join("")}</span>
      </div>
    `);
  }
  document.querySelector("#week-legend").innerHTML = uniqueWaste(state.events).slice(0, 4).map((event) => legendHTML(event)).join("");
}

function renderEmptyWeek() {
  document.querySelector("#week-strip").innerHTML = fullWeekdays.map((_, index) => `
    <div class="week-day"><span class="week-label">${weekdays[index]}</span><span class="week-number">-</span><span class="dots"></span></div>
  `).join("");
}

function renderLegend() {
  document.querySelector("#legend-row").innerHTML = uniqueWaste(state.events).map((event) => legendHTML(event)).join("");
}

function legendHTML(event) {
  return `<span><i class="legend-dot" style="background:${event.color}"></i> ${escapeHTML(event.label)}</span>`;
}

function renderCalendar() {
  const grid = document.querySelector("#calendar-grid");
  const month = state.month;
  document.querySelector("#month-title").textContent = `${monthNames[month.getMonth()]} ${month.getFullYear()}`;
  grid.innerHTML = fullWeekdays.map((day) => `<div class="weekday-cell">${day}</div>`).join("");
  const start = calendarStart(month);
  for (let i = 0; i < 42; i += 1) {
    const date = addDays(start, i);
    const events = eventsForDate(date);
    const muted = date.getMonth() !== month.getMonth() ? "muted" : "";
    grid.insertAdjacentHTML("beforeend", `
      <div class="day-cell ${muted}">
        <span>${date.getDate()}</span>
        ${events.map((event) => {
          const chip = `<i class="dot" style="background:${event.color}"></i>${escapeHTML(event.label)}${event.confidence < 1 ? " *" : ""}`;
          return event.source_url
            ? `<a class="event-chip" href="${event.source_url}" target="_blank" rel="noreferrer">${chip}</a>`
            : `<span class="event-chip">${chip}</span>`;
        }).join("")}
      </div>
    `);
  }
}

function renderUpcoming() {
  const list = document.querySelector("#upcoming-list");
  list.innerHTML = upcomingEvents().slice(0, 9).map((event) => {
    const date = parseDate(event.date);
    return `
      <article class="upcoming-row">
        <div class="date-badge"><strong>${date.getDate()}</strong><small>${monthNames[date.getMonth()].slice(0, 3).toLowerCase()}</small></div>
        <div class="waste-icon" style="background:${event.color}">${wasteGlyph[event.waste_type] || "•"}</div>
        <div class="upcoming-copy">
          <strong>${escapeHTML(event.label)}</strong>
          <span>${event.confidence < 1 ? "Inferat din programul cartierului" : "Sursa oficiala disponibila"}</span>
        </div>
        ${event.source_url ? `<a class="source-arrow" href="${event.source_url}" target="_blank" rel="noreferrer" aria-label="Verifica sursa">↗</a>` : ""}
      </article>
    `;
  }).join("");
}

function renderSources() {
  const source = state.events.find((event) => event.source_url);
  document.querySelector("#source-copy").innerHTML = source
    ? `Datele afisate pastreaza legatura catre sursa oficiala. <a href="${source.source_url}" target="_blank" rel="noreferrer">Verifica la RETIM</a>.`
    : "Calendar generat automat din snapshot-uri RETIM.";
}

function uniqueWaste(events) {
  const seen = new Set();
  return events.filter((event) => {
    if (seen.has(event.waste_type)) return false;
    seen.add(event.waste_type);
    return true;
  });
}

function upcomingEvents() {
  const today = startOfDay(new Date());
  return state.events
    .filter((event) => parseDate(event.date) >= today)
    .sort((a, b) => parseDate(a.date) - parseDate(b.date));
}

function eventsForDate(date) {
  const key = isoDate(date);
  return state.events.filter((event) => event.date === key);
}

function printURL() {
  const month = `${state.month.getFullYear()}-${String(state.month.getMonth() + 1).padStart(2, "0")}`;
  if (staticMode) {
    return pageURL(`print/?place_id=${state.place.id}&month=${month}`);
  }
  return `/print?place_id=${state.place.id}&month=${month}`;
}

function programURL(placeId) {
  if (staticMode) {
    const url = new URL(window.location.href);
    url.searchParams.set("place_id", placeId);
    return url.toString();
  }
  return `/program?place_id=${placeId}`;
}

function pageURL(path) {
  return new URL(path, siteRoot).toString();
}

async function neighborhoodsData() {
  if (staticMode) {
    const catalog = await loadCatalog();
    return catalog.neighborhoods || [];
  }
  const response = await fetch("/api/neighborhoods");
  if (!response.ok) return [];
  const data = await response.json();
  return data.neighborhoods || [];
}

async function searchData(query, neighborhood, limit) {
  if (staticMode) {
    const catalog = await loadCatalog();
    const normalized = normalizeText(query);
    const rows = (catalog.places || [])
      .filter((place) => !neighborhood || place.cartier_norm === neighborhood)
      .map((place) => ({ place, score: scorePlace(normalized, place) }))
      .filter((result) => !normalized || result.score > 0)
      .sort((a, b) => b.score - a.score || a.place.street_raw.localeCompare(b.place.street_raw, "ro"))
      .slice(0, limit);
    return rows;
  }
  const endpoint = neighborhood ? "/api/places" : "/api/search";
  const response = await fetch(`${endpoint}?q=${encodeURIComponent(query)}&cartier_norm=${encodeURIComponent(neighborhood)}&limit=${limit}`);
  const data = await response.json();
  return data.results || [];
}

async function placeData(placeId) {
  if (staticMode) {
    const payload = await fetchJSON(new URL(`places/${placeId}.json`, dataRoot));
    payload.events = (payload.events || []).map(decorateStaticEvent);
    return payload;
  }
  const from = isoDate(monthStart(state.month));
  const to = isoDate(monthEnd(new Date(state.month.getFullYear(), state.month.getMonth() + 2, 1)));
  const response = await fetch(`/api/events?place_id=${placeId}&from=${from}&to=${to}`);
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return response.json();
}

function decorateStaticEvent(event) {
  return {
    ...event,
    label: wasteLabel(event.waste_type),
    color: wasteColor(event.waste_type),
    weekday: fullWeekdays[(parseDate(event.date).getDay() + 6) % 7],
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

function wasteImage(type) {
  return {
    residual: "bin_gray.png",
    bio: "bin_green.png",
    paper: "bin_blue.png",
    plastic_metal: "bin_yellow.png",
    glass: "bin_green.png",
    bulky: "bin_purple.png",
    textile: "bin_pink.png",
    hazardous: "bin_red.png",
  }[type] || "bin_gray.png";
}

async function loadCatalog() {
  if (!state.catalog) {
    state.catalog = await fetchJSON(new URL("catalog.json", dataRoot));
  }
  return state.catalog;
}

async function fetchJSON(url) {
  const response = await fetch(url);
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return response.json();
}

function scorePlace(query, place) {
  if (!query) return 100;
  const candidates = [place.street_norm, `${place.cartier_norm} ${place.street_norm}`, ...(place.aliases || [])];
  let best = 0;
  for (const candidate of candidates) {
    if (!candidate) continue;
    if (candidate === query) best = Math.max(best, 120);
    else if (candidate.includes(query)) best = Math.max(best, 105 - candidate.length + query.length);
    else if (query.includes(candidate)) best = Math.max(best, 92 - query.length + candidate.length);
    else best = Math.max(best, 72 - levenshtein(query, candidate));
  }
  return best < 45 ? 0 : best;
}

function normalizeText(value) {
  return String(value)
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/\b(strada|str|calea|bulevardul|bd|piata|pta|aleea|cartier|cartierul|municipiul|arad)\b/g, " ")
    .replace(/[^a-z0-9]+/g, " ")
    .trim()
    .replace(/\s+/g, " ");
}

function levenshtein(a, b) {
  const previous = Array.from({ length: b.length + 1 }, (_, index) => index);
  for (let i = 1; i <= a.length; i += 1) {
    let diagonal = previous[0];
    previous[0] = i;
    for (let j = 1; j <= b.length; j += 1) {
      const tmp = previous[j];
      previous[j] = Math.min(
        previous[j] + 1,
        previous[j - 1] + 1,
        diagonal + (a[i - 1] === b[j - 1] ? 0 : 1),
      );
      diagonal = tmp;
    }
  }
  return previous[b.length];
}

function downloadICS() {
  const lines = [
    "BEGIN:VCALENDAR",
    "VERSION:2.0",
    "PRODID:-//gunoi-arad//Static GitHub Pages//RO",
    `NAME:Gunoi Arad - ${escapeICS(state.place.street_raw)}`,
    "X-WR-CALNAME:Gunoi Arad",
  ];
  for (const event of state.events) {
    const start = event.date.replaceAll("-", "");
    const end = isoDate(addDays(parseDate(event.date), 1)).replaceAll("-", "");
    lines.push(
      "BEGIN:VEVENT",
      `UID:${start}-${state.place.id}-${event.waste_type}@gunoiarad-static`,
      `SUMMARY:${escapeICS(event.label)} - ${escapeICS(state.place.street_raw)}`,
      `DESCRIPTION:${escapeICS((event.source_url ? `Verifica sursa: ${event.source_url}` : "Program generat din date publice.") + "\n\nMultumiri dezvoltatorului baditaflorin@gmail.com. Mai multe detalii pe: https://baditaflorin.github.io/calendar-ridicare-gunoi-arad/")}`,
      `DTSTART;VALUE=DATE:${start}`,
      `DTEND;VALUE=DATE:${end}`,
      "END:VEVENT",
    );
  }
  lines.push("END:VCALENDAR");
  const blob = new Blob([lines.join("\r\n")], { type: "text/calendar;charset=utf-8" });
  const link = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  link.download = `gunoi-arad-${state.place.id}.ics`;
  link.click();
  URL.revokeObjectURL(link.href);
}

function escapeICS(value) {
  return String(value).replaceAll("\\", "\\\\").replaceAll(",", "\\,").replaceAll(";", "\\;").replaceAll("\n", "\\n");
}

function calendarStart(month) {
  const first = monthStart(month);
  const day = (first.getDay() + 6) % 7;
  return addDays(first, -day);
}

function weekStart(date) {
  const start = startOfDay(date);
  return addDays(start, -((start.getDay() + 6) % 7));
}

function monthStart(date) {
  return new Date(date.getFullYear(), date.getMonth(), 1);
}

function monthEnd(date) {
  return new Date(date.getFullYear(), date.getMonth() + 1, 0);
}

function addDays(date, days) {
  const next = new Date(date);
  next.setDate(next.getDate() + days);
  return next;
}

function parseDate(value) {
  return new Date(`${value}T00:00:00`);
}

function startOfDay(date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

function isoDate(date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

function relativeDate(date) {
  const today = startOfDay(new Date());
  const delta = Math.round((date - today) / 86400000);
  if (delta === 0) return "Astazi";
  if (delta === 1) return "Maine";
  return `${fullWeekdays[(date.getDay() + 6) % 7]}, ${date.getDate()} ${monthNames[date.getMonth()].toLowerCase()}`;
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}
