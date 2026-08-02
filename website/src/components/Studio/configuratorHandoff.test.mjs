import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import {
  CONFIGURATOR_ORIGIN,
  createConfiguratorHandoff,
  isConfiguratorReadyMessage,
  sendStudioConfig,
} from './configuratorHandoff.mjs';

describe('Configurator handoff', () => {
  it('uses the Configurator nonce fragment with no config data in the URL', () => {
    const handoff = createConfiguratorHandoff('yaml', 'version: 4', {
      getRandomValues(bytes) {
        bytes.fill(0xab);
        return bytes;
      },
    });

    assert.equal(handoff.nonce, 'ab'.repeat(32));
    assert.equal(
      handoff.url,
      `https://configurator.ohmyposh.dev/#omp-configurator-nonce=${'ab'.repeat(32)}`,
    );
    assert.equal(handoff.url.includes('version'), false);
  });

  it('only accepts a ready message from the pending configurator window', () => {
    const popupWindow = {};
    const pending = { nonce: 'nonce', popupWindow };
    const message = {
      origin: CONFIGURATOR_ORIGIN,
      source: popupWindow,
      data: { version: 1, type: 'omp-configurator-ready', nonce: 'nonce' },
    };

    assert.equal(isConfiguratorReadyMessage(message, pending), true);
    assert.equal(
      isConfiguratorReadyMessage({ ...message, origin: 'https://example.com' }, pending),
      false,
    );
    assert.equal(isConfiguratorReadyMessage({ ...message, source: {} }, pending), false);
    assert.equal(
      isConfiguratorReadyMessage(
        { ...message, data: { ...message.data, nonce: 'other-nonce' } },
        pending,
      ),
      false,
    );
    assert.equal(
      isConfiguratorReadyMessage(
        { ...message, data: { ...message.data, type: 'omp-studio-config' } },
        pending,
      ),
      false,
    );
    assert.equal(
      isConfiguratorReadyMessage({ ...message, data: { ...message.data, version: 2 } }, pending),
      false,
    );
  });

  it('sends the raw config only to the configurator origin', () => {
    let sentMessage;
    let targetOrigin;
    const pending = {
      nonce: 'nonce',
      format: 'toml',
      text: 'version = 4',
      popupWindow: {
        postMessage(message, origin) {
          sentMessage = message;
          targetOrigin = origin;
        },
      },
    };

    sendStudioConfig(pending);

    assert.deepEqual(sentMessage, {
      version: 1,
      type: 'omp-studio-config',
      nonce: 'nonce',
      format: 'toml',
      text: 'version = 4',
    });
    assert.equal(targetOrigin, CONFIGURATOR_ORIGIN);
  });
});
