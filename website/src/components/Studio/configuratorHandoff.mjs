export const CONFIGURATOR_ORIGIN = 'https://configurator.ohmyposh.dev';
export const CONFIGURATOR_URL = `${CONFIGURATOR_ORIGIN}/`;

const MESSAGE_VERSION = 1;
const READY_MESSAGE_TYPE = 'omp-configurator-ready';
const CONFIG_MESSAGE_TYPE = 'omp-studio-config';
const NONCE_BYTES = 32;
const NONCE_FRAGMENT_KEY = 'omp-configurator-nonce';

export function generateNonce(crypto = globalThis.crypto) {
  if (!crypto || typeof crypto.getRandomValues !== 'function') {
    throw new Error('Secure random values are unavailable.');
  }

  const bytes = new Uint8Array(NONCE_BYTES);
  crypto.getRandomValues(bytes);

  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
}

export function createConfiguratorHandoff(format, text, crypto) {
  const nonce = generateNonce(crypto);
  const url = new URL(CONFIGURATOR_URL);
  url.hash = `${NONCE_FRAGMENT_KEY}=${encodeURIComponent(nonce)}`;

  return { nonce, format, text, url: url.toString() };
}

export function isConfiguratorReadyMessage(event, pending) {
  if (!pending || !pending.popupWindow || event.origin !== CONFIGURATOR_ORIGIN) {
    return false;
  }

  const { data } = event;

  return (
    event.source === pending.popupWindow &&
    data &&
    typeof data === 'object' &&
    data.version === MESSAGE_VERSION &&
    data.type === READY_MESSAGE_TYPE &&
    data.nonce === pending.nonce
  );
}

export function sendStudioConfig(pending) {
  pending.popupWindow.postMessage(
    {
      version: MESSAGE_VERSION,
      type: CONFIG_MESSAGE_TYPE,
      nonce: pending.nonce,
      format: pending.format,
      text: pending.text,
    },
    CONFIGURATOR_ORIGIN,
  );
}
