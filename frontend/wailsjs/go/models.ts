export namespace backend {
	
	export class ExportInfo {
	    skills_count: number;
	    git_cache_count: number;
	    custom_tools_count: number;
	    estimated_size: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skills_count = source["skills_count"];
	        this.git_cache_count = source["git_cache_count"];
	        this.custom_tools_count = source["custom_tools_count"];
	        this.estimated_size = source["estimated_size"];
	    }
	}
	export class FileInfo {
	    name: string;
	    path: string;
	    is_dir: boolean;
	    size: number;
	    // Go type: time
	    modified: any;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.is_dir = source["is_dir"];
	        this.size = source["size"];
	        this.modified = this.convertValues(source["modified"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class GitURLInfo {
	    owner: string;
	    repo: string;
	    sub_path: string;
	    full_url: string;
	    short_ref: string;
	
	    static createFrom(source: any = {}) {
	        return new GitURLInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.owner = source["owner"];
	        this.repo = source["repo"];
	        this.sub_path = source["sub_path"];
	        this.full_url = source["full_url"];
	        this.short_ref = source["short_ref"];
	    }
	}
	export class ImportPreview {
	    version: string;
	    export_date: string;
	    skills_count: number;
	    git_cache_count: number;
	    custom_tools_count: number;
	
	    static createFrom(source: any = {}) {
	        return new ImportPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.export_date = source["export_date"];
	        this.skills_count = source["skills_count"];
	        this.git_cache_count = source["git_cache_count"];
	        this.custom_tools_count = source["custom_tools_count"];
	    }
	}
	export class PathMigrationInfo {
	    has_old_data: boolean;
	    skills_count: number;
	    total_size_mb: number;
	    migration_size_mb: number;
	    old_path: string;
	    new_path: string;
	
	    static createFrom(source: any = {}) {
	        return new PathMigrationInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.has_old_data = source["has_old_data"];
	        this.skills_count = source["skills_count"];
	        this.total_size_mb = source["total_size_mb"];
	        this.migration_size_mb = source["migration_size_mb"];
	        this.old_path = source["old_path"];
	        this.new_path = source["new_path"];
	    }
	}
	export class PathValidationResult {
	    is_valid: boolean;
	    is_writable: boolean;
	    disk_free_gb: number;
	    message: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new PathValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_valid = source["is_valid"];
	        this.is_writable = source["is_writable"];
	        this.disk_free_gb = source["disk_free_gb"];
	        this.message = source["message"];
	        this.status = source["status"];
	    }
	}
	export class SkillsLeaderboardItem {
	    rank: number;
	    name: string;
	    author: string;
	    installs: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillsLeaderboardItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rank = source["rank"];
	        this.name = source["name"];
	        this.author = source["author"];
	        this.installs = source["installs"];
	        this.url = source["url"];
	    }
	}
	export class StorageInfo {
	    total_space: number;
	    used_space: number;
	    free_space: number;
	    skills_count: number;
	    skills_path: string;
	
	    static createFrom(source: any = {}) {
	        return new StorageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_space = source["total_space"];
	        this.used_space = source["used_space"];
	        this.free_space = source["free_space"];
	        this.skills_count = source["skills_count"];
	        this.skills_path = source["skills_path"];
	    }
	}
	export class UpdateInfo {
	    current_version: string;
	    latest_version: string;
	    has_update: boolean;
	    download_url?: string;
	    release_notes?: string;
	    skills_with_update: string[];
	    update_count: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.current_version = source["current_version"];
	        this.latest_version = source["latest_version"];
	        this.has_update = source["has_update"];
	        this.download_url = source["download_url"];
	        this.release_notes = source["release_notes"];
	        this.skills_with_update = source["skills_with_update"];
	        this.update_count = source["update_count"];
	    }
	}

}

export namespace skill {
	
	export class AppSettings {
	    skillhub_path: string;
	    theme: string;
	    auto_update_check: boolean;
	    update_frequency: string;
	    first_run: boolean;
	    custom_categories: string[];
	    github_token: string;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillhub_path = source["skillhub_path"];
	        this.theme = source["theme"];
	        this.auto_update_check = source["auto_update_check"];
	        this.update_frequency = source["update_frequency"];
	        this.first_run = source["first_run"];
	        this.custom_categories = source["custom_categories"];
	        this.github_token = source["github_token"];
	    }
	}
	export class CategoryInfo {
	    name: string;
	    is_preset: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CategoryInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.is_preset = source["is_preset"];
	    }
	}
	export class CustomTool {
	    id: string;
	    name: string;
	    skills_path: string;
	    enabled: boolean;
	    date_added: string;
	
	    static createFrom(source: any = {}) {
	        return new CustomTool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.skills_path = source["skills_path"];
	        this.enabled = source["enabled"];
	        this.date_added = source["date_added"];
	    }
	}
	export class SkillInfo {
	    name: string;
	    author: string;
	    description: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.author = source["author"];
	        this.description = source["description"];
	        this.path = source["path"];
	    }
	}
	export class GitInstallResult {
	    TempPath: string;
	    GitURL: string;
	    Skills: SkillInfo[];
	    InstalledSkills: string[];
	
	    static createFrom(source: any = {}) {
	        return new GitInstallResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TempPath = source["TempPath"];
	        this.GitURL = source["GitURL"];
	        this.Skills = this.convertValues(source["Skills"], SkillInfo);
	        this.InstalledSkills = source["InstalledSkills"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InstallOptions {
	    category: string;
	    tags: string[];
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new InstallOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.tags = source["tags"];
	        this.notes = source["notes"];
	    }
	}
	export class LocalScanResult {
	    TempPath: string;
	    IsZip: boolean;
	    SourcePath: string;
	    Skills: SkillInfo[];
	
	    static createFrom(source: any = {}) {
	        return new LocalScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TempPath = source["TempPath"];
	        this.IsZip = source["IsZip"];
	        this.SourcePath = source["SourcePath"];
	        this.Skills = this.convertValues(source["Skills"], SkillInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SkillStatus {
	    name: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	    }
	}
	export class OperationLog {
	    // Go type: time
	    timestamp: any;
	    action: string;
	    source: string;
	    skills: SkillStatus[];
	    duration_ms: number;
	
	    static createFrom(source: any = {}) {
	        return new OperationLog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.action = source["action"];
	        this.source = source["source"];
	        this.skills = this.convertValues(source["skills"], SkillStatus);
	        this.duration_ms = source["duration_ms"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Skill {
	    id: string;
	    name: string;
	    original_name: string;
	    author: string;
	    description: string;
	    source_type: string;
	    source_url: string;
	    category: string;
	    tags: string[];
	    notes: string;
	    tools_enabled: Record<string, boolean>;
	    // Go type: time
	    installed_at: any;
	    // Go type: time
	    updated_at: any;
	    has_update: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Skill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.original_name = source["original_name"];
	        this.author = source["author"];
	        this.description = source["description"];
	        this.source_type = source["source_type"];
	        this.source_url = source["source_url"];
	        this.category = source["category"];
	        this.tags = source["tags"];
	        this.notes = source["notes"];
	        this.tools_enabled = source["tools_enabled"];
	        this.installed_at = this.convertValues(source["installed_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.has_update = source["has_update"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class ToolInfo {
	    id: string;
	    name: string;
	    skills_path: string;
	    is_installed: boolean;
	    is_enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToolInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.skills_path = source["skills_path"];
	        this.is_installed = source["is_installed"];
	        this.is_enabled = source["is_enabled"];
	    }
	}

}

