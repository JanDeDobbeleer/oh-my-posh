import React from 'react';
import {useLocation} from "react-router-dom";
import CodeBlock from '@theme/CodeBlock';
const queryString = require('query-string');

function Auth() {
  const search = useLocation().search;
  const params = queryString.parse(search);

  if (params.error) {
    let buff = Buffer.from(params.error, 'base64');
    let text = buff.toString('ascii');
    return (
      <div>
        <p>
          Error on authenticating with the <code>{params.segment}</code> API, please provide the following error message
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
    "type": "${params.segment}",
    ...
    "properties": {
      // highlight-start
      "access_token":"${params.access_token}",
      "refresh_token":"${params.refresh_token}",
      "expires_in":"${params.expires_in}"
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
