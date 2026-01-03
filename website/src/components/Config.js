import React from 'react';
import CodeBlock from '@theme/CodeBlock';
import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";
import YAML from 'yaml';
import TOML from 'smol-toml';

function Config(props) {

  const { data, metastring = { json: "", yaml: "", toml: "" } } = props;

  const patchTomlData = () => {
    if (data?.properties) {
      const properties = data.properties;
      delete data.properties;

      return {
        ...data,
        blocks: {
          segments: {
            properties: properties
          }
        }
      };
    }

    return data;
  };

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
        <CodeBlock language="json" metastring={metastring.json}>
          {JSON.stringify(data, null, 2)}
        </CodeBlock>
      </TabItem>
      <TabItem value="yaml">
        <CodeBlock language="yaml" metastring={metastring.yaml}>
          {YAML.stringify(data)}
        </CodeBlock>
      </TabItem>
      <TabItem value="toml">
        <CodeBlock language="toml" metastring={metastring.toml}>
          {TOML.stringify(patchTomlData())}
        </CodeBlock>
      </TabItem>
    </Tabs>
  );
}

export default Config;
