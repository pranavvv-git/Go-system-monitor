const state = {
  timer: null,
  chart: null,
  labels: [],
  cpu: [],
  memory: [],
  disk: [],
};

const el = {
  timestamp: document.querySelector("#timestamp"),
  refresh: document.querySelector("#refresh"),
  cpuUsed: document.querySelector("#cpu-used"),
  cpuCores: document.querySelector("#cpu-cores"),
  cpuProgress: document.querySelector("#cpu-progress"),
  memoryUsed: document.querySelector("#memory-used"),
  memoryFree: document.querySelector("#memory-free"),
  memoryProgress: document.querySelector("#memory-progress"),
  diskUsed: document.querySelector("#disk-used"),
  diskFree: document.querySelector("#disk-free"),
  diskProgress: document.querySelector("#disk-progress"),
  duration: document.querySelector("#duration"),
  durationProgress: document.querySelector("#duration-progress"),
  status: document.querySelector("#status"),
  warning: document.querySelector("#warning"),
  disks: document.querySelector("#disks"),
  metrics: document.querySelector("#metrics"),
  trend: document.querySelector("#trend"),
};

document.body.classList.add("loading");
el.refresh.addEventListener("change", schedule);

initChart();
schedule();
load();

function schedule() {
  if (state.timer) {
    clearInterval(state.timer);
  }
  state.timer = setInterval(load, Number(el.refresh.value));
}

async function load() {
  el.status.textContent = document.body.classList.contains("loading") ? "loading" : "refreshing";
  el.warning.textContent = "";
  try {
    const response = await fetch("/api/stats", { cache: "no-store" });
    if (!response.ok) {
      throw new Error(`stats request failed with ${response.status}`);
    }
    const data = await response.json();
    document.body.classList.remove("loading");
    render(data);
  } catch (error) {
    document.body.classList.remove("loading");
    el.status.textContent = "offline";
    el.warning.textContent = error.message;
  }
}

function render(data) {
  const metrics = data.metrics || [];
  const byName = new Map(metrics.map((metric) => [metricKey(metric), metric]));
  const cpu = value(byName, "gosysmon_cpu_used_ratio");
  const memory = value(byName, "gosysmon_memory_used_ratio");
  const cores = value(byName, "gosysmon_cpu_logical_cores");
  const uptime = value(byName, "gosysmon_system_uptime_seconds");
  const diskRows = disks(metrics);
  const primaryDisk = diskRows[0];

  el.timestamp.textContent = `Collected ${new Date(data.collected_at).toLocaleString()}${uptime == null ? "" : ` - uptime ${duration(uptime)}`}`;
  el.cpuUsed.textContent = cpu == null ? "--" : percent(cpu);
  el.cpuCores.textContent = cores == null ? "-- cores" : `${cores} cores`;
  updateProgress(el.cpuProgress, cpu);

  el.memoryUsed.textContent = memory == null ? "--" : percent(memory);
  el.memoryFree.textContent = `${bytes(value(byName, "gosysmon_memory_available_bytes"))} available`;
  updateProgress(el.memoryProgress, memory);

  el.diskUsed.textContent = primaryDisk ? percent(primaryDisk.used) : "--";
  el.diskFree.textContent = primaryDisk ? `${bytes(primaryDisk.free)} free` : "-- free";
  updateProgress(el.diskProgress, primaryDisk ? primaryDisk.used : null);

  el.duration.textContent = `${data.duration_ms} ms`;
  el.status.textContent = data.warning ? "partial" : "healthy";
  updateProgress(el.durationProgress, Math.min(1, data.duration_ms / 1000));
  el.warning.textContent = data.warning || "";

  pushTrend(cpu, memory, primaryDisk ? primaryDisk.used : null);
  renderDisks(diskRows);
  renderMetrics(metrics);
}

function initChart() {
  if (!window.Chart) {
    el.warning.textContent = "Chart.js could not load; raw metrics are still available.";
    return;
  }

  state.chart = new Chart(el.trend, {
    type: "line",
    data: {
      labels: state.labels,
      datasets: [
        dataset("CPU", state.cpu, "#65a6ff"),
        dataset("Memory", state.memory, "#45d6b5"),
        dataset("Disk", state.disk, "#f0b84d"),
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: { duration: 240 },
      interaction: { mode: "index", intersect: false },
      scales: {
        y: {
          min: 0,
          max: 100,
          ticks: { color: "#9aa6af", callback: (value) => `${value}%` },
          grid: { color: "#313a42" },
          title: { display: true, text: "Usage", color: "#9aa6af" },
        },
        x: {
          ticks: { color: "#9aa6af", maxTicksLimit: 8 },
          grid: { color: "#1f252b" },
          title: { display: true, text: "Sample time", color: "#9aa6af" },
        },
      },
      plugins: {
        legend: { labels: { color: "#f2f5f7" } },
        tooltip: { callbacks: { label: (item) => `${item.dataset.label}: ${item.formattedValue}%` } },
      },
    },
  });
}

function dataset(label, data, color) {
  return {
    label,
    data,
    borderColor: color,
    backgroundColor: `${color}22`,
    borderWidth: 2,
    pointRadius: 2,
    tension: 0.32,
    fill: false,
  };
}

function pushTrend(cpu, memory, disk) {
  state.labels.push(new Date().toLocaleTimeString());
  state.cpu.push(toPercentNumber(cpu));
  state.memory.push(toPercentNumber(memory));
  state.disk.push(toPercentNumber(disk));

  for (const series of [state.labels, state.cpu, state.memory, state.disk]) {
    if (series.length > 48) {
      series.shift();
    }
  }
  if (state.chart) {
    state.chart.update();
  }
}

function metricKey(metric) {
  const labels = metric.labels || {};
  const suffix = Object.keys(labels)
    .sort()
    .map((key) => `${key}=${labels[key]}`)
    .join(",");
  return suffix ? `${metric.name}{${suffix}}` : metric.name;
}

function value(byName, name) {
  const metric = byName.get(name);
  return metric ? metric.value : null;
}

function disks(metrics) {
  const rows = new Map();
  for (const metric of metrics) {
    const path = metric.labels && metric.labels.path;
    if (!path) continue;
    if (!rows.has(path)) rows.set(path, { path });
    const row = rows.get(path);
    if (metric.name === "gosysmon_disk_size_bytes") row.size = metric.value;
    if (metric.name === "gosysmon_disk_free_bytes") row.free = metric.value;
    if (metric.name === "gosysmon_disk_available_bytes") row.available = metric.value;
    if (metric.name === "gosysmon_disk_used_ratio") row.used = metric.value;
  }
  return [...rows.values()].sort((a, b) => a.path.localeCompare(b.path));
}

function renderDisks(rows) {
  el.disks.replaceChildren(
    ...rows.map((row) => {
      const item = document.createElement("div");
      item.className = "disk-row";
      const used = row.used || 0;
      item.innerHTML = `
        <div class="disk-head">
          <strong>${escapeHTML(row.path)}</strong>
          <span>${percent(used)}</span>
        </div>
        <div class="bar"><span class="${level(used)}" style="width:${clampPercent(used)}%"></span></div>
        <div class="disk-path">${bytes(row.free)} free of ${bytes(row.size)}</div>
      `;
      return item;
    })
  );
}

function renderMetrics(metrics) {
  el.metrics.replaceChildren(
    ...metrics.slice(0, 60).map((metric) => {
      const item = document.createElement("div");
      item.className = "metric-row";
      item.innerHTML = `
        <div class="metric-head">
          <strong>${escapeHTML(metricKey(metric))}</strong>
          <span class="metric-value">${formatNumber(metric.value)}</span>
        </div>
        <div class="metric-help">${escapeHTML(metric.help || metric.type)}</div>
      `;
      return item;
    })
  );
}

function updateProgress(node, value) {
  if (!node) return;
  node.className = level(value || 0);
  node.style.width = `${clampPercent(value)}%`;
}

function clampPercent(value) {
  if (value == null || Number.isNaN(value)) return 0;
  return Math.max(0, Math.min(100, value * 100));
}

function toPercentNumber(value) {
  if (value == null || Number.isNaN(value)) return null;
  return Number((value * 100).toFixed(2));
}

function percent(value) {
  return `${(value * 100).toFixed(1)}%`;
}

function level(value) {
  if (value >= 0.9) return "danger";
  if (value >= 0.75) return "warn";
  return "";
}

function bytes(value) {
  if (value == null) return "--";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = value;
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit++;
  }
  return `${size.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`;
}

function duration(seconds) {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}

function formatNumber(value) {
  if (Math.abs(value) >= 1024) return bytes(value);
  if (value >= 0 && value <= 1) return value.toFixed(4);
  return Number(value).toLocaleString();
}

function escapeHTML(value) {
  return String(value).replace(/[&<>"']/g, (char) => {
    return {
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#39;",
    }[char];
  });
}
