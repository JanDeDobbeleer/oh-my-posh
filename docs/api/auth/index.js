const axios = require('axios');

module.exports = async function (context, req) {
  context.log('JavaScript HTTP trigger function processed a request.');
  // https://www.strava.com/oauth/authorize?client_id=76033&response_type=code&redirect_uri=https://ohmyposh.dev/api/auth&approval_prompt=force&scope=read,activity:read

  try {
    const code = (req.query._code || (req.body && req.body.code));
    if (!code) {
      context.res = {
        status: 400
      };
      return;
    }

    const params = {
      client_id: process.env['STRAVA_CLIENT_ID'],
      client_secret: process.env['STRAVA_CLIENT_SECRET'],
      code: code,
      grant_type: 'authorization_code',
    };
    const resp = await axios.post('https://www.strava.com/oauth/token', null, { params: params });

    const body = {
      access_token: resp.data.access_token,
      refresh_token: resp.data.refresh_token,
    }

    context.res = {
      body: body
    };
  } catch (error) {
    context.log(error);
    context.res = {
      body: error,
      status: 500
    };
  }
}
