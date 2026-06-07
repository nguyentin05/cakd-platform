import * as path from 'path';
import { readFileSync, existsSync } from 'fs';
import { fileURLToPath } from 'url';
import * as dotenv from 'dotenv';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

dotenv.config({ path: '../.env' });

import { generateDoc, initLLM } from './core/llm.js';
import { loadManifest, saveManifest } from './core/manifest.js';
import { hasSourceChangedGit } from './core/git.js';
import { hasSourceChangedLocal } from './core/utils.js';

const apiKey = process.env.GEMINI_API_KEY;
if (!apiKey) {
    console.error('Error: GEMINI_API_KEY environment variable is missing.');
    process.exit(1);
}
initLLM(apiKey);

async function main() {
    const projectRoot = path.join(__dirname, '../../');
    const configPath = path.join(projectRoot, 'docs/docs.config.json');
    const forceAll = process.argv.includes('--force');
    const manifest = loadManifest(projectRoot);

    console.log('CAKD Docs Agent V2 starting...');

    if (!existsSync(configPath)) {
        console.error(`Config file not found at ${configPath}`);
        process.exit(1);
    }

    const config = JSON.parse(readFileSync(configPath, 'utf8'));

    for (const module of config.modules) {
        const targetPath = path.join(projectRoot, 'docs', module.target);
        
        const neverGenerated = !manifest.has(module.id);
        
        if (!forceAll && !neverGenerated) {
            const isCI = process.env.CI === 'true';
            const changed = isCI 
                ? hasSourceChangedGit(module, projectRoot)
                : hasSourceChangedLocal(module, projectRoot, targetPath);

            if (!changed) {
                console.log(`Skipping ${module.id} — no source changes`);
                continue;
            }
        }

        const success = await generateDoc(module, projectRoot);
        if (success) {
            manifest.add(module.id);
            saveManifest(projectRoot, manifest);
        }
        
        console.log('  Waiting 15s to respect rate limits...');
        await new Promise(r => setTimeout(r, 15000));
    }

    console.log('Docs Agent V2 finished successfully!');
}

main().catch(console.error);
