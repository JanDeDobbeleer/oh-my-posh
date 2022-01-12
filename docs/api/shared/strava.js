const axios = require('axios');

async function getStravaToken(code) {
  const params = {
    client_id: process.env['STRAVA_CLIENT_ID'],
    client_secret: process.env['STRAVA_CLIENT_SECRET'],
    code: code,
    grant_type: 'authorization_code',
  };
  const resp = await axios.post('https://www.strava.com/api/v3/oauth/token', null, { params: params });

  return {
    access_token: resp.data.access_token,
    refresh_token: resp.data.refresh_token,
    expires_in: resp.data.expires_in
  };
}

async function refreshStravaToken(refresh_token) {
  const params = {
    client_id: process.env['STRAVA_CLIENT_ID'],
    client_secret: process.env['STRAVA_CLIENT_SECRET'],
    refresh_token: refresh_token,
    grant_type: 'refresh_token',
  };
  const resp = await axios.post('https://www.strava.com/api/v3/oauth/token', null, { params: params });

  return {
    access_token: resp.data.access_token,
    refresh_token: resp.data.refresh_token,
    expires_in: resp.data.expires_in
  };
}

module.exports = {
  getStravaToken: getStravaToken,
  refreshStravaToken: refreshStravaToken,
}
