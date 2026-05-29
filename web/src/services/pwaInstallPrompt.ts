export type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed"; platform: string }>;
};

type InstallPromptListener = (prompt: BeforeInstallPromptEvent | null) => void;

let deferredInstallPrompt: BeforeInstallPromptEvent | null = null;
const listeners = new Set<InstallPromptListener>();
let initialized = false;

function notify(): void {
  listeners.forEach((listener) => {
    try {
      listener(deferredInstallPrompt);
    } catch (error) {
      console.error("pwa install prompt listener failed", error);
    }
  });
}

export function initializePWAInstallPromptCapture(): void {
  if (initialized || typeof window === "undefined") {
    return;
  }
  initialized = true;
  window.addEventListener("beforeinstallprompt", (event) => {
    event.preventDefault();
    deferredInstallPrompt = event as BeforeInstallPromptEvent;
    notify();
  });
  window.addEventListener("appinstalled", () => {
    deferredInstallPrompt = null;
    notify();
  });
}

export function subscribePWAInstallPrompt(listener: InstallPromptListener): () => void {
  initializePWAInstallPromptCapture();
  listeners.add(listener);
  listener(deferredInstallPrompt);
  return () => {
    listeners.delete(listener);
  };
}

export function clearDeferredPWAInstallPrompt(): void {
  deferredInstallPrompt = null;
  notify();
}
