import React from 'react';
import CodeBlock from '@theme/CodeBlock';
import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";
import YAML from 'yaml';
import TOML from '@iarna/toml';

function Config(props) {

  const {data} = props;

  return (
    <Tabs
        defaultValue="json"
        groupId="sample"
        values={[
          { label: 'json', value: 'json', },
          { label: 'yaml', value: 'yaml', },
          { label: 'toml', value: 'toml', },
        ]
      }>
      <TabItem value="json">
        <CodeBlock className="language-json">
          {JSON.stringify(data, null, 2)}
        </CodeBlock>
      </TabItem>
      <TabItem value="yaml">
        <CodeBlock className="language-yaml">
          {YAML.stringify(data)}
        </CodeBlock>
      </TabItem>
      <TabItem value="toml">
        <CodeBlock className="language-toml">
          {TOML.stringify(data)}
        </CodeBlock>
      </TabItem>
    </Tabs>
  );
}

export default Config;
