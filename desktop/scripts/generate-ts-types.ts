// This script generates TypeScript types from JSON schemas.
// The output is written to the src/types/generated directory.

import fs from "fs/promises";
import { compile } from "json-schema-to-typescript";
import path from "path";

const SCHEMAS_DIR = path.resolve(process.cwd(), "../docs/public/schemas");
const OUTPUT_DIR = path.resolve(process.cwd(), "src/types/generated");

async function generateTypes() {
  try {
    await fs.mkdir(OUTPUT_DIR, { recursive: true });

    const schemaFiles = await fs.readdir(SCHEMAS_DIR);

    for (const file of schemaFiles) {
      if (!file.endsWith(".json")) continue;

      const schemaPath = path.join(SCHEMAS_DIR, file);
      const schema = JSON.parse(await fs.readFile(schemaPath, "utf-8"));

      const types = await compile(schema, file.replace(".json", ""), {
        bannerComment: `/**
 * This file was automatically generated from ${file}
 * DO NOT MODIFY IT BY HAND
 */`,
        style: {
          singleQuote: true,
          semi: true,
        },
      });

      const outputPath = path.join(
        OUTPUT_DIR,
        `${file.replace(".json", "").replace("_schema", "")}.ts`
      );
      await fs.writeFile(outputPath, types);

      console.log(`Generated types for ${file}`);
    }

    console.log("All ts types generated successfully");
  } catch (error) {
    console.error("Error generating ts types:", error);
    process.exit(1);
  }
}

generateTypes();
