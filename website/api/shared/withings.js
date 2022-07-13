const axios = require('axios');

async function getToken(code) {
  const params = {
    client_id: process.env['WITHINGS_CLIENT_ID'],
    client_secret: process.env['WITHINGS_CLIENT_SECRET'],
    code: code,
    grant_type: 'authorization_code',
    action: 'requesttoken',
    redirect_uri: 'https://ohmyposh.dev/api/auth',
  };

  const resp = await axios.post('https://wbsapi.withings.net/v2/oauth2', null, { params: params });

  if (resp.data.error) {
    throw resp.data.error;
  }

  return {
    access_token: resp.data.body.access_token,
    refresh_token: resp.data.body.refresh_token,
    expires_in: resp.data.body.expires_in
  };
}

async function refreshToken(refresh_token) {
  const params = {
    client_id: process.env['WITHINGS_CLIENT_ID'],
    client_secret: process.env['WITHINGS_CLIENT_SECRET'],
    refresh_token: refresh_token,
    grant_type: 'refresh_token',
    action: 'requesttoken',
    redirect_uri: 'https://ohmyposh.dev/api/auth',
  };
  const resp = await axios.post('https://wbsapi.withings.net/v2/oauth2', null, { params: params });

  return {
    access_token: resp.data.body.access_token,
    refresh_token: resp.data.body.refresh_token,
    expires_in: resp.data.body.expires_in
  };
}

module.exports = {
  getToken: getToken,
  refreshToken: refreshToken,
}
