const axios = require('axios');

async function getToken(code) {
  const params = {
    client_id: process.env['WITHINGS_CLIENT_ID'],
    client_secret: process.env['WITHINGS_CLIENT_SECRET'],
    code: code,
    grant_type: 'authorization_code',
    action: 'requesttoken',
    redirect_uri: 'https://ohmyposh.dev',
  };
  const resp = await axios.post('https://wbsapi.withings.net/v2/oauth2', null, { params: params });

  return {
    access_token: resp.data.access_token,
    refresh_token: resp.data.refresh_token,
    expires_in: resp.data.expires_in
  };
}

async function refreshToken(refresh_token) {
  const params = {
    client_id: process.env['WITHINGS_CLIENT_ID'],
    client_secret: process.env['WITHINGS_CLIENT_SECRET'],
    refresh_token: refresh_token,
    grant_type: 'refresh_token',
    action: 'requesttoken',
    redirect_uri: 'https://ohmyposh.dev',
  };
  const resp = await axios.post('https://wbsapi.withings.net/v2/oauth2', null, { params: params });

  return {
    access_token: resp.data.access_token,
    refresh_token: resp.data.refresh_token,
    expires_in: resp.data.expires_in
  };
}

module.exports = {
  getToken: getToken,
  refreshToken: refreshToken,
}
