# Public Sources

The ETL fetches public pages and stores raw snapshots with hashes before parsing:

- https://retim.ro/utile-arad/zona-1/
- https://retim.ro/utile-arad/campanii-periodice-2/
- https://retim.ro/utile-arad/campanii-periodice-2/anul-3/campania-1/
- https://retim.ro/retim-si-adisigd-arad-anunta-modificarea-programului-de-colectare-pentru-deseurile-reciclabile-din-plastic-si-metal-pubela-galbena-in-municipiul-arad-si-orasele-din-zona-1/
- https://www.primariaarad.ro/dm_arad/portal.nsf/AllByUNID/programul-de-colectare-a-biodeseurilor-si-a-deseurilor-voluminoase-%E2%80%93-primavara-2025--%E2%80%9Caradul-curat%E2%80%9D-000486f2?OpenDocument=
- https://www.primariaarad.ro/dm_arad/portal.nsf/AllByUNID/PROGRAMUL%2B%E2%80%9CARADUL%2BCURAT%E2%80%9D%2B%E2%80%93%2BTOAMNA%2B2025%2BCOLECTAREA%2BGRATUITA%2BA%2BBIODEsEURILOR%2BsI%2BA%2BDEsEURILOR%2BVOLUMINOASE-00125A7E?OpenDocument=
- https://retim.ro/wp-content/uploads/2023/04/grafic-colectare-municipiul-arad.pdf

The current MVP parser treats the RETIM `Zona 1` page and the RETIM `Campania 1 / Anul 3` page as primary sources. Other sources are registered as fetch targets for audit and future parser expansion.
