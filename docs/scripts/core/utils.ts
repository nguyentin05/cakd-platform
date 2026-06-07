import { readFileSync, statSync, existsSync } from 'fs';
import * as path from 'path';
import { globSync } from 'glob';

export function gatherCodeContext(patterns: string[], cwd: string): string {
    let context = '';
    for (const pattern of patterns) {
        const files = globSync(pattern, { cwd });
        for (const file of files) {
            const fullPath = path.join(cwd, file);
            try {
                const content = readFileSync(fullPath, 'utf8');
                context += `\n--- File: ${file} ---\n${content}\n`;
            } catch (e) {
                console.warn(`Could not read ${fullPath}`);
            }
        }
    }
    return context;
}

export function hasSourceChangedLocal(
    moduleConfig: any, 
    projectRoot: string,
    targetPath: string
): boolean {
    if (!existsSync(targetPath)) return true;
    
    const targetStat = statSync(targetPath);
    
    for (const pattern of moduleConfig.sources) {
        const files = globSync(pattern, { cwd: projectRoot });
        for (const file of files) {
            const sourceStat = statSync(path.join(projectRoot, file));
            if (sourceStat.mtimeMs > targetStat.mtimeMs) {
                console.log(`  Source changed: ${file}`);
                return true;
            }
        }
    }
    return false;
}
