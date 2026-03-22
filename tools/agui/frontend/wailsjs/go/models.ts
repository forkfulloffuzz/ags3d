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

