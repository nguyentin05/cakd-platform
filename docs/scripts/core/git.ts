import { execSync } from 'child_process';
import { globSync } from 'glob';

export function hasSourceChangedGit(moduleConfig: any, projectRoot: string): boolean {
    try {
        const changedFiles = execSync('git diff --name-only HEAD~1 HEAD', { 
            cwd: projectRoot 
        }).toString().trim().split('\n');
        
        for (const pattern of moduleConfig.sources) {
            const files = globSync(pattern, { cwd: projectRoot });
            if (files.some(f => changedFiles.includes(f))) return true;
        }
        return false;
    } catch {
        return true; 
    }
}
