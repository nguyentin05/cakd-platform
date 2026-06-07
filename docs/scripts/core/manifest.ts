import { readFileSync, writeFileSync, existsSync } from 'fs';
import * as path from 'path';

export function loadManifest(projectRoot: string): Set<string> {
    const manifestPath = path.join(projectRoot, 'docs/.docs-manifest.json');
    if (!existsSync(manifestPath)) return new Set();
    try {
        const data = JSON.parse(readFileSync(manifestPath, 'utf8'));
        return new Set(data.generated);
    } catch {
        return new Set();
    }
}

export function saveManifest(projectRoot: string, generated: Set<string>) {
    const manifestPath = path.join(projectRoot, 'docs/.docs-manifest.json');
    writeFileSync(manifestPath, JSON.stringify({ 
        generated: [...generated] 
    }, null, 2));
}
