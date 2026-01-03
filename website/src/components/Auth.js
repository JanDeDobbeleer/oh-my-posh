import React from 'react';
import {useLocation} from '@docusaurus/router';
import CodeBlock from '@theme/CodeBlock';

function Auth() {
  const search = useLocation().search;
  const params = new URLSearchParams(search);

  if (params.get('error')) {
    let buff = Buffer.from(params.get('error'), 'base64');
    let text = buff.toString('ascii');
    return (
      <div>
        <p>
          Error on authenticating with the <code>{params.get('segment')}</code> API, please provide the following error message
          in a <a href='https://github.com/JanDeDobbeleer/oh-my-posh/issues/new/choose'>ticket</a> in
          case this was unexpected.
        </p>
        <CodeBlock className="language-jsstacktrace">
          {text}
        </CodeBlock>
      </div>
    );
  }

  const config = `
  {
    "type": "${params.get('segment')}",
    ...
    "options": {
      // highlight-start
      "access_token": "${params.get('access_token')}",
      "refresh_token": "${params.get('refresh_token')}",
      "expires_in": ${params.get('expires_in')}
      // highlight-end
    }
  }
  `;

  return (
    <div>
      <p>
        Use the following snippet to adjust your segment and enable the authentication.
      </p>
      <CodeBlock className="language-json" title="config.omp.json">
        {config}
      </CodeBlock>
    </div>
  );
}

export default Auth;
