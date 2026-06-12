import { GoogleGenerativeAI } from '@google/generative-ai';
import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'fs';
import * as path from 'path';
import { gatherCodeContext } from './utils.js';

let model: any = null;

export function initLLM(apiKey: string) {
    const genAI = new GoogleGenerativeAI(apiKey);
    model = genAI.getGenerativeModel({ 
        model: 'gemini-2.5-flash',
        generationConfig: {
            temperature: 0.0,
        }
    });
}

export async function callWithRetry(prompt: string, maxRetries = 3): Promise<string> {
    for (let i = 0; i < maxRetries; i++) {
        try {
            const result = await model.generateContent(prompt);
            return result.response.text();
        } catch (error) {
            if (i === maxRetries - 1) throw error;
            console.warn(`  Retry ${i + 1}/${maxRetries}...`);
            await new Promise(r => setTimeout(r, 2000 * (i + 1)));
        }
    }
    throw new Error('Max retries exceeded');
}

export function stripMarkdownWrapper(text: string): string {
    return text
        .replace(/^```(?:markdown|md)?\n/, '')
        .replace(/\n```$/, '')
        .trim();
}

function ensureDir(filePath: string) {
    const dirname = path.dirname(filePath);
    if (!existsSync(dirname)) {
        mkdirSync(dirname, { recursive: true });
    }
}

function parseFrontmatterAndBody(content: string): { frontmatter: string; body: string } {
    if (!content.startsWith('---')) {
        return { frontmatter: '', body: content };
    }
    const parts = content.split('---');
    if (parts.length < 3) {
        return { frontmatter: '', body: content };
    }
    const frontmatter = parts[1].trim();
    const body = parts.slice(2).join('---').trim();
    return { frontmatter, body };
}

export async function generateDoc(moduleConfig: any, projectRoot: string): Promise<boolean> {
    const targetPath = path.join(projectRoot, 'docs', moduleConfig.target);
    const templatePath = path.join(projectRoot, 'docs', moduleConfig.promptTemplate);
    
    console.log(`Processing module: ${moduleConfig.id} -> ${targetPath}`);

    const sourceContext = gatherCodeContext(moduleConfig.sources, projectRoot);
    if (!sourceContext) {
        console.warn(`No source files found for module ${moduleConfig.id}`);
        return false;
    }

    let promptTemplate = '';
    try {
        promptTemplate = readFileSync(templatePath, 'utf8');
    } catch (e) {
        console.error(`Failed to read prompt template: ${templatePath}`);
        return false;
    }

    let existingDoc = '';
    if (existsSync(targetPath)) {
        existingDoc = readFileSync(targetPath, 'utf8');
    }

    const styleGuidePath = path.join(projectRoot, 'docs/prompts/_style-guide.md');
    const styleGuide = existsSync(styleGuidePath) 
        ? readFileSync(styleGuidePath, 'utf8') 
        : '';

    let finalPrompt = `${styleGuide}\n\n---\n\n${promptTemplate}\n\n`;
    if (existingDoc) {
        finalPrompt += `
=== EXISTING DOCUMENTATION (read carefully before writing) ===
${existingDoc}
=== END EXISTING DOCUMENTATION ===

CRITICAL INSTRUCTIONS:
1. Return the COMPLETE updated document, not just changed sections
2. Preserve ALL existing content that is not directly affected by the source code changes
3. Do not change section order, headings, or frontmatter
4. Do not rephrase or rewrite sections that are still accurate
5. Only update/add content that reflects actual changes in the source code below

`;
    }
    finalPrompt += `=== SOURCE CODE ===\n${sourceContext}\n=== END SOURCE CODE ===\n`;

    try {
        console.log(`  Sending request to Gemini...`);
        let text = await callWithRetry(finalPrompt);
        
        text = stripMarkdownWrapper(text);

        if (existingDoc) {
            const original = parseFrontmatterAndBody(existingDoc);
            if (original.frontmatter) {
                const generated = parseFrontmatterAndBody(text);
                text = `---\n${original.frontmatter}\n---\n\n${generated.body}`;
            }
        }

        ensureDir(targetPath);
        writeFileSync(targetPath, text);
        console.log(`  Saved ${targetPath}`);
        return true;
    } catch (error) {
        console.error(`  Failed to generate ${targetPath}:`, error);
        return false;
    }
}
