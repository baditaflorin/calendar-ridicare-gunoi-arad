document.addEventListener("DOMContentLoaded", async () => {
  const neighborhood = document.querySelector("#map-neighborhood");
  const street = document.querySelector("#map-street");
  const results = document.querySelector("#map-results");
  const staticMode = window.GUNOI_STATIC === true;
  const siteRoot = new URL("../", document.currentScript.src);
  const dataRoot = new URL("data/", siteRoot);
  let catalog;

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
      const query = encodeURIComponent(`Strada ${place.street_raw}, ${place.cartier}, Arad, Romania`);
      return `
        <article class="map-result">
          <div>
            <strong>${escapeHTML(place.street_raw)}</strong>
            <span>${escapeHTML(place.cartier || "Arad")}</span>
          </div>
          <a href="${staticMode ? new URL(`program/?place_id=${place.id}`, siteRoot).toString() : `/program?place_id=${place.id}`}">Program</a>
          <a href="https://www.openstreetmap.org/search?query=${query}" target="_blank" rel="noreferrer">OSM</a>
        </article>
      `;
    }).join("") || `<p class="quiet">Nu am gasit strazi pentru filtrul curent.</p>`;
  };

  neighborhood.addEventListener("change", loadPlaces);
  street.addEventListener("input", debounce(loadPlaces, 180));
  loadPlaces();
});

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
