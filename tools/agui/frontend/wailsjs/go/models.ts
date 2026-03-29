export namespace api {
	
	export class Vec3 {
	    x: number;
	    y: number;
	    z: number;
	
	    static createFrom(source: any = {}) {
	        return new Vec3(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.z = source["z"];
	    }
	}
	export class ParsedBlocker {
	    name: string;
	    position: Vec3;
	    size: Vec3;
	
	    static createFrom(source: any = {}) {
	        return new ParsedBlocker(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.position = this.convertValues(source["position"], Vec3);
	        this.size = this.convertValues(source["size"], Vec3);
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
	export class ParsedCamera {
	    name: string;
	    position: Vec3;
	    lookAt: Vec3;
	
	    static createFrom(source: any = {}) {
	        return new ParsedCamera(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.position = this.convertValues(source["position"], Vec3);
	        this.lookAt = this.convertValues(source["lookAt"], Vec3);
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
	export class ParsedChar {
	    name: string;
	    displayName?: string;
	    type: string;
	    mesh?: string;
	    animations?: Record<string, string>;
	    spriteSheet?: string;
	    spriteAngles?: number;
	    framesPerAngle?: number;
	    frameSize?: number[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ParsedChar(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	        this.type = source["type"];
	        this.mesh = source["mesh"];
	        this.animations = source["animations"];
	        this.spriteSheet = source["spriteSheet"];
	        this.spriteAngles = source["spriteAngles"];
	        this.framesPerAngle = source["framesPerAngle"];
	        this.frameSize = source["frameSize"];
	        this.error = source["error"];
	    }
	}
	export class ParsedHotspot {
	    name: string;
	    position: Vec3;
	    size: Vec3;
	
	    static createFrom(source: any = {}) {
	        return new ParsedHotspot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.position = this.convertValues(source["position"], Vec3);
	        this.size = this.convertValues(source["size"], Vec3);
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
	export class ParsedPoint {
	    name: string;
	    position: Vec3;
	
	    static createFrom(source: any = {}) {
	        return new ParsedPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.position = this.convertValues(source["position"], Vec3);
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
	export class ParsedSpawnPoint {
	    name: string;
	    character?: string;
	    position: Vec3;
	
	    static createFrom(source: any = {}) {
	        return new ParsedSpawnPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.character = source["character"];
	        this.position = this.convertValues(source["position"], Vec3);
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
	export class Vec2 {
	    x: number;
	    z: number;
	
	    static createFrom(source: any = {}) {
	        return new Vec2(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.z = source["z"];
	    }
	}
	export class ParsedWalkable {
	    name: string;
	    position: Vec3;
	    size: Vec2;
	
	    static createFrom(source: any = {}) {
	        return new ParsedWalkable(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.position = this.convertValues(source["position"], Vec3);
	        this.size = this.convertValues(source["size"], Vec2);
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
	export class ParsedRoom {
	    name: string;
	    initialCamera?: string;
	    cameras?: ParsedCamera[];
	    points?: ParsedPoint[];
	    walkableSurfaces?: ParsedWalkable[];
	    blockerVolumes?: ParsedBlocker[];
	    spawnPoints?: ParsedSpawnPoint[];
	    hotspots?: ParsedHotspot[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ParsedRoom(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.initialCamera = source["initialCamera"];
	        this.cameras = this.convertValues(source["cameras"], ParsedCamera);
	        this.points = this.convertValues(source["points"], ParsedPoint);
	        this.walkableSurfaces = this.convertValues(source["walkableSurfaces"], ParsedWalkable);
	        this.blockerVolumes = this.convertValues(source["blockerVolumes"], ParsedBlocker);
	        this.spawnPoints = this.convertValues(source["spawnPoints"], ParsedSpawnPoint);
	        this.hotspots = this.convertValues(source["hotspots"], ParsedHotspot);
	        this.error = source["error"];
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
	
	
	export class ValidateIssue {
	    file: string;
	    severity: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ValidateIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.severity = source["severity"];
	        this.message = source["message"];
	    }
	}
	export class ValidateResult {
	    issues: ValidateIssue[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ValidateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.issues = this.convertValues(source["issues"], ValidateIssue);
	        this.error = source["error"];
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
	

}

export namespace main {
	
	export class ProjectInfo {
	    root: string;
	    name: string;
	    startRoom: string;
	    startCharacter: string;
	    renderingMode: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.name = source["name"];
	        this.startRoom = source["startRoom"];
	        this.startCharacter = source["startCharacter"];
	        this.renderingMode = source["renderingMode"];
	        this.error = source["error"];
	    }
	}
	export class RefFolderInfo {
	    root: string;
	    name: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new RefFolderInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.name = source["name"];
	        this.error = source["error"];
	    }
	}
	export class SourceFile {
	    path: string;
	    rel: string;
	    ext: string;
	
	    static createFrom(source: any = {}) {
	        return new SourceFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.rel = source["rel"];
	        this.ext = source["ext"];
	    }
	}
	export class TranspileResult {
	    source: string;
	    tokens: string;
	    astText: string;
	    astDot: string;
	    symbols: string;
	    symDot: string;
	    blocking: string;
	    gdscript: string;
	    emitView: string;
	    sourceMap: any[][];
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new TranspileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.tokens = source["tokens"];
	        this.astText = source["astText"];
	        this.astDot = source["astDot"];
	        this.symbols = source["symbols"];
	        this.symDot = source["symDot"];
	        this.blocking = source["blocking"];
	        this.gdscript = source["gdscript"];
	        this.emitView = source["emitView"];
	        this.sourceMap = source["sourceMap"];
	        this.errors = source["errors"];
	    }
	}

}

