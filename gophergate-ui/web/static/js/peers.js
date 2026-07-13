(() => {
  const initCreatePeerDialog = () => {
    const dialog = document.getElementById("create-peer-dialog");

    if (!dialog) {
      return;
    }

    document.querySelectorAll("[data-create-peer-open]").forEach((button) => {
      button.addEventListener("click", () => {
        if (typeof dialog.showModal === "function") {
          dialog.showModal();
          return;
        }

        dialog.setAttribute("open", "open");
      });
    });

    document.querySelectorAll("[data-create-peer-close]").forEach((button) => {
      button.addEventListener("click", () => {
        dialog.close();
      });
    });

    dialog.addEventListener("click", (event) => {
      if (event.target === dialog) {
        dialog.close();
      }
    });

    if (dialog.dataset.openOnLoad === "true" && !dialog.open) {
      dialog.showModal();
    }
  };

  const initDeletePeerDialog = () => {
    const dialog = document.getElementById("delete-peer-dialog");

    if (!dialog) {
      return;
    }

    const publicKeyInput = document.getElementById("delete-peer-public-key");
    const peerName = document.getElementById("delete-peer-name");
    const peerKey = document.getElementById("delete-peer-key");

    if (!publicKeyInput || !peerName || !peerKey) {
      return;
    }

    document.querySelectorAll("[data-delete-peer]").forEach((button) => {
      button.addEventListener("click", () => {
        publicKeyInput.value = button.dataset.publicKey || "";
        peerName.textContent = button.dataset.peerName || "this peer";
        peerKey.textContent =
          button.dataset.publicKeyShort ||
          button.dataset.publicKey ||
          "—";

        if (typeof dialog.showModal === "function") {
          dialog.showModal();
          return;
        }

        dialog.setAttribute("open", "open");
      });
    });

    document.querySelectorAll("[data-delete-cancel]").forEach((button) => {
      button.addEventListener("click", () => {
        dialog.close();
      });
    });

    dialog.addEventListener("click", (event) => {
      if (event.target === dialog) {
        dialog.close();
      }
    });
  };

  const init = () => {
    initCreatePeerDialog();
    initDeletePeerDialog();
  };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();