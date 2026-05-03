const state = {
  place: null,
  events: [],
  month: new Date(),
  neighborhoods: [],
};

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
  const params = new URLSearchParams(window.location.search);
  const placeId = params.get("place_id");
  bindSearch();
  bindButtons();
  renderEmptyWeek();
  loadNeighborhoods();
  if (placeId) {
    loadPlace(placeId);
  } else {
    const input = document.querySelector("#street-search");
    input.value = "Densuseanu";
    search(input.value, true);
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
  document.querySelector("#print-button").addEventListener("click", () => {
    if (state.place) window.open(printURL(), "_blank", "noopener");
  });
  document.querySelector("#share-button").addEventListener("click", async () => {
    if (!state.place) return;
    const url = `${window.location.origin}/program?place_id=${state.place.id}`;
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
  const endpoint = neighborhood ? "/api/places" : "/api/search";
  const response = await fetch(`${endpoint}?q=${encodeURIComponent(query)}&cartier_norm=${encodeURIComponent(neighborhood)}&limit=12`);
  const data = await response.json();
  const results = data.results || [];
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
  const response = await fetch("/api/neighborhoods");
  if (!response.ok) return;
  const data = await response.json();
  state.neighborhoods = data.neighborhoods || [];
  const select = document.querySelector("#neighborhood-select");
  select.insertAdjacentHTML("beforeend", state.neighborhoods.map((item) => (
    `<option value="${escapeHTML(item.norm)}">${escapeHTML(item.name)} (${item.count})</option>`
  )).join(""));
  const total = state.neighborhoods.reduce((sum, item) => sum + item.count, 0);
  document.querySelector("#catalog-count").textContent = `${total} strazi reale din ${state.neighborhoods.length} cartiere`;
}

async function loadPlace(placeId) {
  const from = isoDate(monthStart(state.month));
  const to = isoDate(monthEnd(new Date(state.month.getFullYear(), state.month.getMonth() + 2, 1)));
  const response = await fetch(`/api/events?place_id=${placeId}&from=${from}&to=${to}`);
  if (!response.ok) return;
  const data = await response.json();
  state.place = data.place;
  state.events = data.events || [];
  document.querySelector("#street-search").value = `${state.place.street_raw}, ${state.place.cartier || "Arad"}`;
  window.history.replaceState(null, "", `/program?place_id=${state.place.id}`);
  renderAll();
}

function loadEventsForCurrentWindow() {
  if (state.place) loadPlace(state.place.id);
}

function renderAll() {
  const place = state.place;
  document.querySelector("#place-title").textContent = `Programul pentru Strada ${place.street_raw}, Arad`;
  document.querySelector("#place-subtitle").textContent = place.cartier ? `Cartier ${place.cartier}` : "Municipiul Arad";
  document.querySelector("#freshness").textContent = "actualizat";
  document.querySelector("#print-link").href = printURL();
  document.querySelector("#ics-link").href = `/ics?place_id=${place.id}`;
  renderNext();
  renderWeek();
  renderLegend();
  renderCalendar();
  renderUpcoming();
  renderSources();
}

function renderNext() {
  const next = upcomingEvents()[0];
  const card = document.querySelector("#next-card");
  if (!next) {
    card.innerHTML = `<div><p>Nu exista ridicari in intervalul incarcat.</p></div>`;
    return;
  }
  const date = parseDate(next.date);
  card.style.setProperty("--bin-color", next.color);
  card.innerHTML = `
    <div>
      <p class="eyebrow">Urmatoarea colectare</p>
      <h3>${relativeDate(date)}</h3>
      <p><strong>${escapeHTML(next.label)}</strong></p>
      <p>Incepand cu ${next.start_time || "07:00"}</p>
    </div>
    <div class="bin-visual" aria-hidden="true"></div>
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
        ${events.map((event) => `<span class="event-chip"><i class="dot" style="background:${event.color}"></i>${escapeHTML(event.label)}</span>`).join("")}
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
        <div class="upcoming-copy"><strong>${escapeHTML(event.label)}</strong><span>Scoate pubela pana la ${event.start_time || "07:00"}</span></div>
      </article>
    `;
  }).join("");
}

function renderSources() {
  const source = state.events.find((event) => event.source_url);
  document.querySelector("#source-copy").innerHTML = source
    ? `Ultima sursa verificata: <a href="${source.source_url}" target="_blank" rel="noreferrer">RETIM</a>.`
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
  return `/print?place_id=${state.place.id}&month=${month}`;
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
