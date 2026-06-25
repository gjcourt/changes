"use strict";

// Chromatic tonics offered in the Key selector. Spelling per key is decided
// server-side by the theory engine; these are just the 12 targets.
const KEYS = ["C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"];

const els = {
  tune: document.getElementById("tune"),
  key: document.getElementById("key"),
  roman: document.getElementById("roman"),
  reset: document.getElementById("reset"),
  sheet: document.getElementById("sheet"),
  meta: document.getElementById("meta"),
};

let originalTonic = "C";

// tonic strips a trailing minor marker from a key name ("Cm" -> "C").
function tonic(key) {
  return key.replace(/m(in)?$/, "");
}

async function getJSON(url) {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return res.json();
}

async function loadTunes() {
  const tunes = await getJSON("/api/standards");
  els.tune.innerHTML = "";
  for (const t of tunes) {
    const opt = document.createElement("option");
    opt.value = t.id;
    opt.textContent = `${t.title} — ${t.composer} (${t.key})`;
    els.tune.appendChild(opt);
  }
  for (const k of KEYS) {
    const opt = document.createElement("option");
    opt.value = k;
    opt.textContent = k;
    els.key.appendChild(opt);
  }
}

async function render() {
  const id = els.tune.value;
  const key = els.key.value;
  const roman = els.roman.checked;
  const data = await getJSON(
    `/api/standards/${encodeURIComponent(id)}?key=${encodeURIComponent(key)}&roman=${roman ? 1 : 0}`
  );
  draw(data);
}

function draw(data) {
  els.sheet.innerHTML = "";
  for (const section of data.sections) {
    const wrap = document.createElement("section");
    wrap.className = "chart-section";

    const label = document.createElement("div");
    label.className = "section-label";
    label.textContent = section.label;
    wrap.appendChild(label);

    const grid = document.createElement("div");
    grid.className = "bars";
    for (const bar of section.bars) {
      const cell = document.createElement("div");
      cell.className = "bar";
      for (const chord of bar) {
        const c = document.createElement("div");
        c.className = "chord";
        const sym = document.createElement("span");
        sym.className = "symbol";
        sym.textContent = chord.symbol;
        c.appendChild(sym);
        if (chord.roman) {
          const rn = document.createElement("span");
          rn.className = "roman";
          rn.textContent = chord.roman;
          c.appendChild(rn);
        }
        cell.appendChild(c);
      }
      grid.appendChild(cell);
    }
    wrap.appendChild(grid);
    els.sheet.appendChild(wrap);
  }
  els.meta.textContent =
    `${data.title} · ${data.composer} · original key ${data.originalKey} · ` +
    `now in ${data.key} · ${data.form} · ${data.meter}`;
}

function setKeyToOriginal() {
  els.key.value = tonic(originalTonic);
}

async function onTuneChange() {
  const opt = els.tune.selectedOptions[0];
  // The original key is embedded in the option label "(Key)".
  const m = opt && opt.textContent.match(/\(([^)]+)\)\s*$/);
  originalTonic = m ? m[1] : "C";
  setKeyToOriginal();
  await render();
}

async function main() {
  try {
    await loadTunes();
    els.tune.addEventListener("change", onTuneChange);
    els.key.addEventListener("change", render);
    els.roman.addEventListener("change", render);
    els.reset.addEventListener("click", async () => {
      setKeyToOriginal();
      await render();
    });
    await onTuneChange();
  } catch (err) {
    els.sheet.innerHTML = `<p class="error">Failed to load: ${err.message}</p>`;
  }
}

main();
