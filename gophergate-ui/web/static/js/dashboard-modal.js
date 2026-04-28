document.addEventListener("DOMContentLoaded", () => {
  const backdrop = document.getElementById("peer-modal-backdrop");
  const closeBtn = document.getElementById("peer-modal-close");

  const titleEl = document.getElementById("peer-modal-title");
  const totalRxEl = document.getElementById("peer-total-rx");
  const totalTxEl = document.getElementById("peer-total-tx");
  const publicKeyEl = document.getElementById("peer-public-key");
  const allowedIpsEl = document.getElementById("peer-allowed-ips");
  const chartCanvas = document.getElementById("peer-traffic-chart");

  if (
    !backdrop ||
    !closeBtn ||
    !titleEl ||
    !totalRxEl ||
    !totalTxEl ||
    !publicKeyEl ||
    !allowedIpsEl ||
    !chartCanvas
  ) {
    return;
  }

  let trafficChart = null;

  function humanBytes(n) {
    const unit = 1024;

    if (!Number.isFinite(n) || n < 0) {
      return "0 B";
    }

    if (n < unit) {
      return `${Math.floor(n)} B`;
    }

    let div = unit;
    let exp = 0;
    let v = Math.floor(n / unit);

    while (v >= unit) {
      div *= unit;
      exp++;
      v = Math.floor(v / unit);
    }

    const value = n / div;
    const suffixes = ["KB", "MB", "GB", "TB", "PB"];
    const safeExp = Math.min(exp, suffixes.length - 1);

    if (value >= 100) {
      return `${value.toFixed(0)} ${suffixes[safeExp]}`;
    }
    if (value >= 10) {
      return `${value.toFixed(1)} ${suffixes[safeExp]}`;
    }
    return `${value.toFixed(2)} ${suffixes[safeExp]}`;
  }

  function maskPublicKey(publicKey, startChars = 8, endChars = 8) {
  if (!publicKey) return "—";
  if (publicKey.length <= startChars + endChars) return publicKey;
  
  const maskedStart = "X".repeat(startChars);
  const middle = publicKey.slice(startChars, -endChars);
  const maskedEnd = "X".repeat(endChars);

  return `${maskedStart}${middle}${maskedEnd}`;
}

  function openModal() {
    backdrop.classList.remove("hidden");
    backdrop.classList.add("flex");
    document.body.classList.add("overflow-hidden");
  }

  function closeModal() {
    backdrop.classList.add("hidden");
    backdrop.classList.remove("flex");
    document.body.classList.remove("overflow-hidden");
  }

  function setLoadingState(peerName, publicKey) {
    titleEl.textContent = peerName || "Peer";
    totalRxEl.textContent = "Loading...";
    totalTxEl.textContent = "Loading...";
    publicKeyEl.textContent = maskPublicKey(publicKey) || "—";
    allowedIpsEl.textContent = "Loading...";
  }

  function toDeltas(values) {
    if (!Array.isArray(values) || values.length === 0) {
      return [];
    }

    const deltas = [0];

    for (let i = 1; i < values.length; i++) {
      const current = Number(values[i] || 0);
      const previous = Number(values[i - 1] || 0);

      deltas.push(Math.max(0, current - previous));
    }

    return deltas;
  }

  function destroyTrafficChart() {
    if (trafficChart) {
      trafficChart.destroy();
      trafficChart = null;
    }
  }

  function renderTrafficChart(points) {
    destroyTrafficChart();

    if (!Array.isArray(points) || points.length === 0) {
      const ctx = chartCanvas.getContext("2d");
      ctx.clearRect(0, 0, chartCanvas.width, chartCanvas.height);
      return;
    }

    const labels = points.map((p) => {
      const date = new Date(p.timestamp);
      if (Number.isNaN(date.getTime())) {
        return p.timestamp;
      }
      return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    });

    const rxRaw = points.map((p) => Number(p.rx_bytes || 0));
    const txRaw = points.map((p) => Number(p.tx_bytes || 0));

    const rxData = toDeltas(rxRaw);
    const txData = toDeltas(txRaw);

    trafficChart = new Chart(chartCanvas, {
      type: "line",
      data: {
        labels,
        datasets: [
          {
            label: "Rx",
            data: rxData,
            borderColor: "#5bc2f0",
            backgroundColor: "transparent",
            borderWidth: 3,
            tension: 0.35,
            pointRadius: 0
          },
          {
            label: "Tx",
            data: txData,
            borderColor: "#f6d6be",
            backgroundColor: "transparent",
            borderWidth: 3,
            tension: 0.35,
            pointRadius: 0
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          mode: "index",
          intersect: false
        },
        plugins: {
          legend: {
            display: false
          },
          tooltip: {
            callbacks: {
              label(context) {
                return `${context.dataset.label}: ${humanBytes(context.raw)}`;
              }
            }
          }
        },
        scales: {
          x: {
            ticks: {
              color: "rgba(202,211,245,0.65)",
              maxRotation: 0,
              autoSkip: true
            },
            grid: {
              color: "rgba(91,194,240,0.08)"
            }
          },
          y: {
            ticks: {
              color: "rgba(202,211,245,0.65)",
              callback(value) {
                return humanBytes(Number(value));
              }
            },
            grid: {
              color: "rgba(91,194,240,0.08)"
            }
          }
        }
      }
    });
  }

  async function loadPeerModal(publicKey, peerName) {
    let loadingShown = false;

    const loadingTimer = setTimeout(() => {
      setLoadingState(peerName, publicKey);
      openModal();
      loadingShown = true;
    }, 180);

    try {
      const response = await fetch(
        `/dashboard/peers/${encodeURIComponent(publicKey)}/modal`
      );

      if (!response.ok) {
        throw new Error("Failed to load peer modal data");
      }

      const data = await response.json();

      clearTimeout(loadingTimer);

      titleEl.textContent = data.name || peerName || "Peer";
      totalRxEl.textContent = humanBytes(Number(data.total_rx || 0));
      totalTxEl.textContent = humanBytes(Number(data.total_tx || 0));
      publicKeyEl.textContent = maskPublicKey(data.public_key || publicKey);
      allowedIpsEl.textContent =
        Array.isArray(data.allowed_ips) && data.allowed_ips.length > 0
          ? data.allowed_ips.join(", ")
          : "—";

      renderTrafficChart(data.points || []);

      if (!loadingShown) {
        openModal();
      }
    } catch (error) {
      clearTimeout(loadingTimer);
      console.error(error);

      setLoadingState(peerName, publicKey);
      totalRxEl.textContent = "Error";
      totalTxEl.textContent = "Error";
      allowedIpsEl.textContent = "Unable to load";
      destroyTrafficChart();

      if (!loadingShown) {
        openModal();
      }
    }
  }

  document.querySelectorAll(".peer-row").forEach((row) => {
    row.addEventListener("click", () => {
      const publicKey = row.dataset.publicKey;
      const peerName = row.dataset.peerName;

      if (!publicKey) return;
      loadPeerModal(publicKey, peerName);
    });

    row.addEventListener("keydown", (event) => {
      if (event.key === "Enter" || event.key === " ") {
        event.preventDefault();

        const publicKey = row.dataset.publicKey;
        const peerName = row.dataset.peerName;

        if (!publicKey) return;
        loadPeerModal(publicKey, peerName);
      }
    });
  });

  closeBtn.addEventListener("click", closeModal);

  backdrop.addEventListener("click", (event) => {
    if (event.target === backdrop) {
      closeModal();
    }
  });

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && !backdrop.classList.contains("hidden")) {
      closeModal();
    }
  });
});