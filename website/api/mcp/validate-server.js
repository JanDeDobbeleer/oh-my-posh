const Ajv = require("ajv");
const addFormats = require("ajv-formats");
const fs = require("fs");
const https = require("https");

const ajv = new Ajv({ strict: false, allErrors: true });
addFormats(ajv);

// Download schema if not present
const schemaPath = "server.schema.json";
const schemaUrl = "https://static.modelcontextprotocol.io/schemas/2025-10-17/server.schema.json";

function downloadSchema() {
  return new Promise((resolve, reject) => {
    https.get(schemaUrl, (res) => {
      let data = "";
      res.on("data", (chunk) => data += chunk);
      res.on("end", () => {
        fs.writeFileSync(schemaPath, data);
        resolve(JSON.parse(data));
      });
    }).on("error", reject);
  });
}

async function validateServer() {
  let schema;
  
  // Load or download schema
  if (fs.existsSync(schemaPath)) {
    schema = JSON.parse(fs.readFileSync(schemaPath, "utf8"));
  } else {
    console.log("Downloading schema...");
    schema = await downloadSchema();
  }

  // Load server.json
  const serverJson = JSON.parse(fs.readFileSync("server.json", "utf8"));

  // Validate
  const validate = ajv.compile(schema);
  const valid = validate(serverJson);

  if (valid) {
    console.log("✅ server.json is valid!");
    console.log(`   Name: ${serverJson.name}`);
    console.log(`   Version: ${serverJson.version}`);
    console.log(`   Transport: ${serverJson.remotes?.[0]?.type || 'N/A'}`);
    process.exit(0);
  } else {
    console.error("❌ Validation errors:");
    validate.errors.forEach(err => {
      console.error(`   ${err.instancePath}: ${err.message}`);
      if (err.params) {
        console.error(`   Details: ${JSON.stringify(err.params)}`);
      }
    });
    process.exit(1);
  }
}

validateServer().catch(err => {
  console.error("Error:", err.message);
  process.exit(1);
});
