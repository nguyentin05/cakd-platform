import { GoogleGenerativeAI } from '@google/generative-ai';
import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'fs';
import * as path from 'path';
import { globSync } from 'glob';
import * as dotenv from 'dotenv';

dotenv.config({ path: '../.env' });

const apiKey = process.env.GEMINI_API_KEY;
if (!apiKey) {
    console.error('Error: GEMINI_API_KEY environment variable is missing.');
    process.exit(1);
}

const genAI = new GoogleGenerativeAI(apiKey);
const model = genAI.getGenerativeModel({ model: 'gemini-1.5-flash' });

function gatherCodeContext(pattern: string, cwd: string): string {
    const files = globSync(pattern, { cwd });
    let context = '';
    for (const file of files) {
        const fullPath = path.join(cwd, file);
        try {
            const content = readFileSync(fullPath, 'utf8');
            context += `\n--- File: ${file} ---\n${content}\n`;
        } catch (e) {
            console.warn(`Could not read ${fullPath}`);
        }
    }
    return context;
}

function ensureDir(filePath: string) {
    const dirname = path.dirname(filePath);
    if (!existsSync(dirname)) {
        mkdirSync(dirname, { recursive: true });
    }
}

async function generateDoc(promptContext: string, systemInstruction: string, outputPath: string) {
    console.log(`Generating docs for ${outputPath}...`);
    try {
        const prompt = `${systemInstruction}\n\nHere is the source code context:\n${promptContext}`;
        const result = await model.generateContent(prompt);
        const text = result.response.text();
        
        ensureDir(outputPath);
        writeFileSync(outputPath, text);
        console.log(`Saved ${outputPath}`);
    } catch (error) {
        console.error(`Failed to generate ${outputPath}:`, error);
    }
}

async function main() {
    const projectRoot = path.join(__dirname, '../../');
    const docsContentDir = path.join(__dirname, '../src/content/docs');

    console.log('CAKD Docs Agent starting...');

    // 1. Generate CLI Reference
    const cmdContext = gatherCodeContext('cmd/**/*.go', projectRoot);
    if (cmdContext) {
        await generateDoc(
            cmdContext,
            `You are an expert technical writer. Write a markdown file documenting the CLI commands found in this Go code. 
Must include Astro Starlight frontmatter at the very top:
---
title: CLI Reference
description: Command line interface reference.
---

Include all flags, commands, and a brief description.`,
            path.join(docsContentDir, 'reference/cli.md')
        );
    }

    const internalContext = gatherCodeContext('internal/**/*.go', projectRoot);
    if (internalContext) {
        await generateDoc(
            internalContext,
            `You are an expert software architect. Write a markdown file explaining the architecture of this project based on the Go code provided.
Must include Astro Starlight frontmatter at the very top:
---
title: Architecture
description: Internal architecture of the CAKD Platform.
---

Focus on how the components (like template engine, terraform bridge) interact. You can use Mermaid diagrams if necessary.`,
            path.join(docsContentDir, 'explanation/architecture.md')
        );
    }

    console.log('Docs Agent finished successfully!');
}

main().catch(console.error);
