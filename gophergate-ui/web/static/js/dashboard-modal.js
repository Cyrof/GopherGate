const backdrop = document.getElementById("peer-modal-backdrop");
const modal = document.getElementById("peer-modal");
const closeBtn = document.getElementById("peer-modal-close");

const titleEl = document.getElementById("peer-modal-title");
const totalRxEl = document.getElementById("peer-total-rx");
const totalTxEl = document.getElementById("peer-total-tx");
const publicKeyEl = document.getElementById("peer-public-key");
const allowedIpsEl = document.getElementById("peer-allowed-ips");

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
