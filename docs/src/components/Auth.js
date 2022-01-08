import React from 'react';
import {useLocation} from "react-router-dom";
import CodeBlock from '@theme/CodeBlock';

function Auth() {
  const search = useLocation().search;
  const segment = new URLSearchParams(search).get('segment');
  const access_token = new URLSearchParams(search).get('access_token');
  const refresh_token = new URLSearchParams(search).get('refresh_token');

  const config = `
  {
    "type": "${segment}",
    ...
    "properties": {
      "access_token":"${access_token}",
      "refresh_token":"${refresh_token}"
    }
  }
  `;

  return (
    <CodeBlock className="language-json" title="config.omp.json">
      {config}
    </CodeBlock>
  );
}

export default Auth;
