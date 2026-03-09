import { html, useState, useEffect } from '../lib.js';
import { useLang } from '../i18n.js';
import { api } from '../api.js';

export function ConfigureDialog({ instanceName, onClose, onConfigure }) {
  const { t } = useLang();
  const [provider, setProvider] = useState('anthropic');
  const [apiKey, setApiKey] = useState('');
  const [model, setModel] = useState('');
  const [channel, setChannel] = useState('');
  const [channelToken, setChannelToken] = useState('');
  const [configuring, setConfiguring] = useState(false);
  const [loading, setLoading] = useState(true);
  const [apiKeyHint, setApiKeyHint] = useState('');
  const [channelTokenHint, setChannelTokenHint] = useState('');

  useEffect(() => {
    api.getConfigStatus(instanceName).then((status) => {
      if (status && status.configured) {
        if (status.provider) setProvider(status.provider);
        if (status.model) {
          const parts = status.model.split('/');
          setModel(parts.length > 1 ? parts.slice(1).join('/') : status.model);
        }
        if (status.channel) setChannel(status.channel);
        if (status.api_key_hint) setApiKeyHint(status.api_key_hint);
        if (status.channel_token_hint) setChannelTokenHint(status.channel_token_hint);
      }
    }).catch(() => {}).finally(() => setLoading(false));
  }, [instanceName]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setConfiguring(true);
    try {
      await onConfigure(instanceName, { provider, api_key: apiKey, model, channel, channel_token: channelToken });
    } finally {
      setConfiguring(false);
    }
  };

  return html`
    <div class="dialog-overlay" onClick=${onClose}>
      <div class="dialog" onClick=${(e) => e.stopPropagation()}>
        <div class="dialog-header">
          <h2>${t('configure.title')}: ${instanceName}</h2>
          <button class="dialog-close" onClick=${onClose}>✕</button>
        </div>
        ${loading ? html`
          <div class="dialog-body"><p>${t('dashboard.loading')}</p></div>
        ` : html`
          <form onSubmit=${handleSubmit}>
            <div class="dialog-body">
              <label class="form-label">
                ${t('configure.provider')}
                <select class="form-input" value=${provider} onChange=${(e) => setProvider(e.target.value)}>
                  <option value="anthropic">Anthropic</option>
                  <option value="openai">OpenAI</option>
                  <option value="google">Google</option>
                </select>
              </label>
              <label class="form-label">
                ${t('configure.apiKey')}
                <input
                  type="password"
                  class="form-input"
                  value=${apiKey}
                  onInput=${(e) => setApiKey(e.target.value)}
                  placeholder=${apiKeyHint || t('configure.apiKeyHint')}
                  required
                  autofocus
                />
              </label>
              <label class="form-label">
                ${t('configure.model')}
                <input
                  type="text"
                  class="form-input"
                  value=${model}
                  onInput=${(e) => setModel(e.target.value)}
                  placeholder=${t('configure.modelHint')}
                />
              </label>
              <label class="form-label">
                ${t('configure.channel')}
                <select class="form-input" value=${channel} onChange=${(e) => setChannel(e.target.value)}>
                  <option value="">-- None --</option>
                  <option value="telegram">Telegram</option>
                  <option value="discord">Discord</option>
                  <option value="slack">Slack</option>
                </select>
              </label>
              ${channel && html`
                <label class="form-label">
                  ${t('configure.channelToken')}
                  <input
                    type="password"
                    class="form-input"
                    value=${channelToken}
                    onInput=${(e) => setChannelToken(e.target.value)}
                    placeholder=${channelTokenHint || t('configure.channelTokenHint')}
                  />
                </label>
              `}
            </div>
            <div class="dialog-footer">
              <button type="button" class="btn btn-ghost" onClick=${onClose}>${t('configure.cancel')}</button>
              <button type="submit" class="btn btn-primary" disabled=${configuring || !apiKey}>
                ${configuring ? t('configure.configuring') : t('configure.submit')}
              </button>
            </div>
          </form>
        `}
      </div>
    </div>
  `;
}
