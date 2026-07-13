const initCreatePeerDialog = () => {
  const dialog = document.getElementById("create-peer-dialog");

  if (!dialog) {
    return;
  }

  document.querySelectorAll("[data-create-peer-open]").forEach((button) => {
    button.addEventListener("click", () => {
      if (!dialog.open) {
        dialog.showModal();
      }
    });
  });

  document.querySelectorAll("[data-create-peer-close]").forEach((button) => {
    button.addEventListener("click", () => {
      if (dialog.open) {
        dialog.close();
      }
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