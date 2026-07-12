(() => {
  const initPeerDeleteDialog = () => {
    const dialog = document.getElementById("delete-peer-dialog");

    if (!dialog) {
      return;
    }

    const publicKeyInput = document.getElementById("delete-peer-public-key");
    const peerName = document.getElementById("delete-peer-name");
    const peerKey = document.getElementById("delete-peer-key");

    const openDialog = (button) => {
      const name = button.dataset.peerName || "this peer";
      const publicKey = button.dataset.publicKey || "";
      const shortKey = button.dataset.publicKeyShort || publicKey || "—";

      publicKeyInput.value = publicKey;
      peerName.textContent = name;
      peerKey.textContent = shortKey;

      if (!dialog.open) {
        dialog.showModal();
      }
    };

    const closeDialog = () => {
      if (dialog.open) {
        dialog.close();
      }

      publicKeyInput.value = "";
    };

    document.querySelectorAll("[data-delete-peer]").forEach((button) => {
      button.addEventListener("click", () => openDialog(button));
    });

    document.querySelectorAll("[data-delete-cancel]").forEach((button) => {
      button.addEventListener("click", closeDialog);
    });

    dialog.addEventListener("click", (event) => {
      if (event.target === dialog) {
        closeDialog();
      }
    });

    document.addEventListener("keydown", (event) => {
      if (event.key === "Escape" && dialog.open) {
        closeDialog();
      }
    });
  };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initPeerDeleteDialog);
  } else {
    initPeerDeleteDialog();
  }
})();