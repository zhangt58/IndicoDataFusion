export namespace backend {
	
	export class AbstractData {
	    title: string;
	    author: string;
	    description: string;
	    keywords: string[];
	
	    static createFrom(source: any = {}) {
	        return new AbstractData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.author = source["author"];
	        this.description = source["description"];
	        this.keywords = source["keywords"];
	    }
	}
	export class ContributionData {
	    title: string;
	    contributor: string;
	    type: string;
	    status: string;
	    submittedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ContributionData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.contributor = source["contributor"];
	        this.type = source["type"];
	        this.status = source["status"];
	        this.submittedAt = source["submittedAt"];
	    }
	}
	export class Event {
	    id: string;
	    title: string;
	    description: string;
	    // Go type: time
	    startDate?: any;
	    // Go type: time
	    endDate?: any;
	    location?: string;
	    address?: string;
	    category?: string;
	
	    static createFrom(source: any = {}) {
	        return new Event(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.startDate = this.convertValues(source["startDate"], null);
	        this.endDate = this.convertValues(source["endDate"], null);
	        this.location = source["location"];
	        this.address = source["address"];
	        this.category = source["category"];
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

