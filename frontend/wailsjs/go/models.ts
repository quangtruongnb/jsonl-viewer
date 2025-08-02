export namespace main {
	
	export class FileStats {
	    totalLines: number;
	    validRecords: number;
	    invalidLines: number[];
	    commonFields: string[];
	    fileSize: number;
	
	    static createFrom(source: any = {}) {
	        return new FileStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalLines = source["totalLines"];
	        this.validRecords = source["validRecords"];
	        this.invalidLines = source["invalidLines"];
	        this.commonFields = source["commonFields"];
	        this.fileSize = source["fileSize"];
	    }
	}
	export class HighlightMatch {
	    text: string;
	    startPos: number;
	    endPos: number;
	    fieldName: string;
	
	    static createFrom(source: any = {}) {
	        return new HighlightMatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.startPos = source["startPos"];
	        this.endPos = source["endPos"];
	        this.fieldName = source["fieldName"];
	    }
	}
	export class JSONLFile {
	    name: string;
	    path: string;
	    size: number;
	    records: number;
	    // Go type: time
	    loadedAt: any;
	    // Go type: time
	    modifiedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new JSONLFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.records = source["records"];
	        this.loadedAt = this.convertValues(source["loadedAt"], null);
	        this.modifiedAt = this.convertValues(source["modifiedAt"], null);
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
	export class JSONRecord {
	    lineNumber: number;
	    content: {[key: string]: any};
	    rawJSON: string;
	
	    static createFrom(source: any = {}) {
	        return new JSONRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lineNumber = source["lineNumber"];
	        this.content = source["content"];
	        this.rawJSON = source["rawJSON"];
	    }
	}
	export class PaginatedRecords {
	    records: JSONRecord[];
	    offset: number;
	    limit: number;
	    total: number;
	    hasMore: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PaginatedRecords(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.records = this.convertValues(source["records"], JSONRecord);
	        this.offset = source["offset"];
	        this.limit = source["limit"];
	        this.total = source["total"];
	        this.hasMore = source["hasMore"];
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
	export class SearchOptions {
	    query: string;
	    caseSensitive: boolean;
	    useLucene: boolean;
	    selectedField: string;
	    offset: number;
	    limit: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.caseSensitive = source["caseSensitive"];
	        this.useLucene = source["useLucene"];
	        this.selectedField = source["selectedField"];
	        this.offset = source["offset"];
	        this.limit = source["limit"];
	    }
	}
	export class SearchResult {
	    records: JSONRecord[];
	    offset: number;
	    limit: number;
	    total: number;
	    totalMatches: number;
	    hasMore: boolean;
	    query: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.records = this.convertValues(source["records"], JSONRecord);
	        this.offset = source["offset"];
	        this.limit = source["limit"];
	        this.total = source["total"];
	        this.totalMatches = source["totalMatches"];
	        this.hasMore = source["hasMore"];
	        this.query = source["query"];
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

