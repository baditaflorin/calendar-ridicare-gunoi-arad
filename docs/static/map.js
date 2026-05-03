const staticMode = window.GUNOI_STATIC === true;
const siteRoot = new URL("../", document.currentScript.src);
const dataRoot = new URL("data/", siteRoot);

document.addEventListener("DOMContentLoaded", async () => {
  const neighborhood = document.querySelector("#map-neighborhood");
  const street = document.querySelector("#map-street");
  const results = document.querySelector("#map-results");
  let catalog;

  // Initialize Leaflet map
  const map = L.map('map').setView([46.1866, 21.3123], 13);
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
  }).addTo(map);

  let currentPopup = null;

  const data = staticMode ? await fetchJSON(new URL("catalog.json", dataRoot)) : await fetchJSON("/api/neighborhoods");
  catalog = data;
  const neighborhoods = data.neighborhoods || [];
  neighborhood.insertAdjacentHTML("beforeend", neighborhoods.map((item) => (
    `<option value="${escapeHTML(item.norm)}">${escapeHTML(item.name)} (${item.count})</option>`
  )).join(""));

  const loadPlaces = async () => {
    const params = new URLSearchParams({
      cartier_norm: neighborhood.value,
      q: street.value,
      limit: "18",
    });
    if (!neighborhood.value && !street.value.trim()) {
      results.innerHTML = `<p class="quiet">Alege un cartier sau cauta o strada.</p>`;
      return;
    }
    const rows = staticMode ? staticPlaces(catalog, neighborhood.value, street.value, 18) : (await fetchJSON(`/api/places?${params.toString()}`)).results || [];
    results.innerHTML = rows.map(({ place }) => {
      return `
        <article class="map-result">
          <div>
            <strong>${escapeHTML(place.street_raw)}</strong>
            <span>${escapeHTML(place.cartier || "Arad")}</span>
          </div>
          <button class="map-pin-btn" data-street="${escapeHTML(place.street_raw)}" data-cartier="${escapeHTML(place.cartier || "")}" data-id="${place.id}">📍 Vezi</button>
          <a href="${staticMode ? new URL(`program/?place_id=${place.id}`, siteRoot).toString() : `/program?place_id=${place.id}`}">Program</a>
        </article>
      `;
    }).join("") || `<p class="quiet">Nu am gasit strazi pentru filtrul curent.</p>`;
  };

  results.addEventListener("click", async (e) => {
    const btn = e.target.closest(".map-pin-btn");
    if (!btn) return;
    const street = btn.dataset.street;
    const cartier = btn.dataset.cartier;
    const placeId = btn.dataset.id;
    const query = `Strada ${street}, ${cartier ? cartier + ',' : ''} Arad, Romania`;
    
    btn.textContent = "⌛";
    btn.disabled = true;
    try {
      const res = await fetch(`https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(query)}&limit=1`);
      const data = await res.json();
      if (data && data.length > 0) {
        const item = data[0];
        const lat = parseFloat(item.lat);
        const lon = parseFloat(item.lon);
        map.flyTo([lat, lon], 16);
        const programLink = staticMode ? new URL(`program/?place_id=${placeId}`, siteRoot).toString() : `/program?place_id=${placeId}`;
        if (currentPopup) currentPopup.remove();
        currentPopup = L.popup()
          .setLatLng([lat, lon])
          .setContent(`<div style="text-align:center;"><strong>${escapeHTML(street)}</strong><br>${escapeHTML(cartier)}<br><br><a href="${programLink}" style="font-weight:bold;">Vezi Programul</a></div>`)
          .openOn(map);
      } else {
        alert("Nu am gasit coordonatele pentru strada cautata.");
      }
    } catch(err) {
      console.error(err);
      alert("Eroare la conectarea cu OpenStreetMap.");
    } finally {
      btn.textContent = "📍 Vezi";
      btn.disabled = false;
    }
  });

  neighborhood.addEventListener("change", loadPlaces);
  street.addEventListener("input", debounce(loadPlaces, 180));
  
  document.querySelector("#btn-pattern-today").addEventListener("click", async () => {
    const targetNeighborhood = neighborhood.value || "centru";
    const streets = (catalog.places || []).filter(p => p.cartier_norm === targetNeighborhood);
    
    if (streets.length === 0) {
      alert("Nu am gasit strazi in cartierul selectat.");
      return;
    }

    const today = new Date().toISOString().slice(0, 10);
    const results = document.querySelector("#map-results");
    results.innerHTML = `<p>Analizam ${streets.length} strazi din ${targetNeighborhood}... va dura putin din cauza geocodarii.</p>`;

    for (const place of streets) {
      try {
        // 1. Fetch events
        const payload = await fetchJSON(new URL(`places/${place.id}.json`, dataRoot));
        const eventToday = (payload.events || []).find(e => e.date === today);
        
        if (eventToday) {
          const query = `Strada ${place.street_raw}, Arad, Romania`;
          const res = await fetch(`https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(query)}&limit=1`);
          const geoData = await res.json();
          
          if (geoData && geoData.length > 0) {
            const lat = parseFloat(geoData[0].lat);
            const lon = parseFloat(geoData[0].lon);
            const color = wasteColor(eventToday.waste_type);
            
            L.circleMarker([lat, lon], {
              radius: 8,
              fillColor: color,
              color: "#fff",
              weight: 2,
              opacity: 1,
              fillOpacity: 0.8
            }).addTo(map)
              .bindPopup(`<strong>${escapeHTML(place.street_raw)}</strong><br>Azi: ${escapeHTML(wasteLabel(eventToday.waste_type))}`);
          }
          // Delay to respect Nominatim rate limit
          await new Promise(r => setTimeout(r, 1000));
        }
      } catch (err) {
        console.error(`Failed to analyze ${place.street_raw}`, err);
      }
    }
    results.innerHTML = `<p>Analiza finalizata pentru ${targetNeighborhood}.</p>`;
  });

  loadPlaces();
});

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

function staticPlaces(catalog, neighborhood, query, limit) {
  const q = normalizeText(query);
  return (catalog.places || [])
    .filter((place) => !neighborhood || place.cartier_norm === neighborhood)
    .map((place) => ({ place, score: scorePlace(q, place) }))
    .filter((result) => !q || result.score > 0)
    .sort((a, b) => b.score - a.score || a.place.street_raw.localeCompare(b.place.street_raw, "ro"))
    .slice(0, limit);
}

function scorePlace(query, place) {
  if (!query) return 100;
  const candidates = [place.street_norm, `${place.cartier_norm} ${place.street_norm}`, ...(place.aliases || [])];
  let best = 0;
  for (const candidate of candidates) {
    if (!candidate) continue;
    if (candidate === query) best = Math.max(best, 120);
    else if (candidate.includes(query)) best = Math.max(best, 105 - candidate.length + query.length);
    else best = Math.max(best, 70 - Math.abs(candidate.length - query.length));
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

async function fetchJSON(url) {
  const response = await fetch(url);
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return response.json();
}

function debounce(fn, delay) {
  let timer;
  return () => {
    clearTimeout(timer);
    timer = setTimeout(fn, delay);
  };
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}
