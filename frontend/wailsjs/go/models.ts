export namespace backend {
	
	export class CustomField {
	    id: number;
	    name: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new CustomField(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.value = source["value"];
	    }
	}
	export class Review {
	
	
	    static createFrom(source: any = {}) {
	        return new Review(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class ContribType {
	    id: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ContribType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class Track {
	    code: string;
	    id: number;
	    title: string;
	
	    static createFrom(source: any = {}) {
	        return new Track(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.id = source["id"];
	        this.title = source["title"];
	    }
	}
	export class Submitter {
	    affiliation: string;
	    email: string;
	    first_name: string;
	    last_name: string;
	    full_name: string;
	    avatar_url: string;
	    id: number;
	    identifier: string;
	
	    static createFrom(source: any = {}) {
	        return new Submitter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.affiliation = source["affiliation"];
	        this.email = source["email"];
	        this.first_name = source["first_name"];
	        this.last_name = source["last_name"];
	        this.full_name = source["full_name"];
	        this.avatar_url = source["avatar_url"];
	        this.id = source["id"];
	        this.identifier = source["identifier"];
	    }
	}
	export class Person {
	    affiliation: string;
	    email: string;
	    author_type: string;
	    first_name: string;
	    last_name: string;
	    is_speaker: boolean;
	    person_id: number;
	
	    static createFrom(source: any = {}) {
	        return new Person(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.affiliation = source["affiliation"];
	        this.email = source["email"];
	        this.author_type = source["author_type"];
	        this.first_name = source["first_name"];
	        this.last_name = source["last_name"];
	        this.is_speaker = source["is_speaker"];
	        this.person_id = source["person_id"];
	    }
	}
	export class Judge {
	    affiliation: string;
	    email: string;
	    first_name: string;
	    last_name: string;
	    full_name: string;
	    avatar_url: string;
	    id: number;
	    identifier: string;
	
	    static createFrom(source: any = {}) {
	        return new Judge(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.affiliation = source["affiliation"];
	        this.email = source["email"];
	        this.first_name = source["first_name"];
	        this.last_name = source["last_name"];
	        this.full_name = source["full_name"];
	        this.avatar_url = source["avatar_url"];
	        this.id = source["id"];
	        this.identifier = source["identifier"];
	    }
	}
	export class AbstractData {
	    id: number;
	    friendly_id: number;
	    state: string;
	    title: string;
	    content: string;
	    score?: number;
	    judge?: Judge;
	    judgment_comment: string;
	    judgment_dt: string;
	    persons: Person[];
	    submitter?: Submitter;
	    accepted_track?: Track;
	    accepted_contrib_type?: ContribType;
	    submitted_contrib_type?: ContribType;
	    reviewed_for_tracks: Track[];
	    submitted_for_tracks: Track[];
	    reviews: Review[];
	    comments: any[];
	    custom_fields: CustomField[];
	    submitted_dt: string;
	    modified_dt: string;
	    modified_by?: Submitter;
	    submission_comment: string;
	    duplicate_of?: number;
	    merged_into?: number;
	    files: any[];
	
	    static createFrom(source: any = {}) {
	        return new AbstractData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.friendly_id = source["friendly_id"];
	        this.state = source["state"];
	        this.title = source["title"];
	        this.content = source["content"];
	        this.score = source["score"];
	        this.judge = this.convertValues(source["judge"], Judge);
	        this.judgment_comment = source["judgment_comment"];
	        this.judgment_dt = source["judgment_dt"];
	        this.persons = this.convertValues(source["persons"], Person);
	        this.submitter = this.convertValues(source["submitter"], Submitter);
	        this.accepted_track = this.convertValues(source["accepted_track"], Track);
	        this.accepted_contrib_type = this.convertValues(source["accepted_contrib_type"], ContribType);
	        this.submitted_contrib_type = this.convertValues(source["submitted_contrib_type"], ContribType);
	        this.reviewed_for_tracks = this.convertValues(source["reviewed_for_tracks"], Track);
	        this.submitted_for_tracks = this.convertValues(source["submitted_for_tracks"], Track);
	        this.reviews = this.convertValues(source["reviews"], Review);
	        this.comments = source["comments"];
	        this.custom_fields = this.convertValues(source["custom_fields"], CustomField);
	        this.submitted_dt = source["submitted_dt"];
	        this.modified_dt = source["modified_dt"];
	        this.modified_by = this.convertValues(source["modified_by"], Submitter);
	        this.submission_comment = source["submission_comment"];
	        this.duplicate_of = source["duplicate_of"];
	        this.merged_into = source["merged_into"];
	        this.files = source["files"];
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
	export class ConfigPathInfo {
	    path: string;
	    fromEnv: boolean;
	    envVarName: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigPathInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.fromEnv = source["fromEnv"];
	        this.envVarName = source["envVarName"];
	    }
	}
	export class TestConfig {
	    dataDir: string;
	    eventInfo: string;
	    abstracts: string;
	    contribs: string;
	
	    static createFrom(source: any = {}) {
	        return new TestConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dataDir = source["dataDir"];
	        this.eventInfo = source["eventInfo"];
	        this.abstracts = source["abstracts"];
	        this.contribs = source["contribs"];
	    }
	}
	export class IndicoConfig {
	    baseUrl: string;
	    eventId: number;
	    apiToken: string;
	    timeout?: string;
	
	    static createFrom(source: any = {}) {
	        return new IndicoConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.baseUrl = source["baseUrl"];
	        this.eventId = source["eventId"];
	        this.apiToken = source["apiToken"];
	        this.timeout = source["timeout"];
	    }
	}
	export class DataSource {
	    name: string;
	    type: string;
	    indico?: IndicoConfig;
	    test?: TestConfig;
	
	    static createFrom(source: any = {}) {
	        return new DataSource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.indico = this.convertValues(source["indico"], IndicoConfig);
	        this.test = this.convertValues(source["test"], TestConfig);
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
	export class ConfigDataUI {
	    activeDataSourceName: string;
	    dataSources: DataSource[];
	    pathInfo: ConfigPathInfo;
	
	    static createFrom(source: any = {}) {
	        return new ConfigDataUI(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.activeDataSourceName = source["activeDataSourceName"];
	        this.dataSources = this.convertValues(source["dataSources"], DataSource);
	        this.pathInfo = this.convertValues(source["pathInfo"], ConfigPathInfo);
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
	
	
	export class Folder {
	    _type: string;
	    id: number;
	    title?: string;
	    description: string;
	    attachments: any[];
	    default_folder: boolean;
	    is_protected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Folder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this._type = source["_type"];
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.attachments = source["attachments"];
	        this.default_folder = source["default_folder"];
	        this.is_protected = source["is_protected"];
	    }
	}
	export class ContributionParticipation {
	    _type: string;
	    _fossil: string;
	    first_name: string;
	    last_name: string;
	    fullName: string;
	    id: string;
	    affiliation: string;
	    emailHash: string;
	    db_id: number;
	    person_id: number;
	    email: string;
	
	    static createFrom(source: any = {}) {
	        return new ContributionParticipation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this._type = source["_type"];
	        this._fossil = source["_fossil"];
	        this.first_name = source["first_name"];
	        this.last_name = source["last_name"];
	        this.fullName = source["fullName"];
	        this.id = source["id"];
	        this.affiliation = source["affiliation"];
	        this.emailHash = source["emailHash"];
	        this.db_id = source["db_id"];
	        this.person_id = source["person_id"];
	        this.email = source["email"];
	    }
	}
	export class DateInfo {
	    date: string;
	    time: string;
	    tz: string;
	
	    static createFrom(source: any = {}) {
	        return new DateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.time = source["time"];
	        this.tz = source["tz"];
	    }
	}
	export class ContributionData {
	    _type: string;
	    _fossil: string;
	    id: string;
	    db_id: number;
	    friendly_id: number;
	    title: string;
	    description: string;
	    startDate: DateInfo;
	    endDate: DateInfo;
	    duration: number;
	    location: string;
	    room: string;
	    roomFullname: string;
	    type: string;
	    session: string;
	    track: string;
	    speakers: ContributionParticipation[];
	    primaryauthors: ContributionParticipation[];
	    coauthors: ContributionParticipation[];
	    keywords: any[];
	    references: any[];
	    board_number: string;
	    code: string;
	    url: string;
	    note: any;
	    material: any[];
	    folders: Folder[];
	    allowed: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ContributionData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this._type = source["_type"];
	        this._fossil = source["_fossil"];
	        this.id = source["id"];
	        this.db_id = source["db_id"];
	        this.friendly_id = source["friendly_id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.startDate = this.convertValues(source["startDate"], DateInfo);
	        this.endDate = this.convertValues(source["endDate"], DateInfo);
	        this.duration = source["duration"];
	        this.location = source["location"];
	        this.room = source["room"];
	        this.roomFullname = source["roomFullname"];
	        this.type = source["type"];
	        this.session = source["session"];
	        this.track = source["track"];
	        this.speakers = this.convertValues(source["speakers"], ContributionParticipation);
	        this.primaryauthors = this.convertValues(source["primaryauthors"], ContributionParticipation);
	        this.coauthors = this.convertValues(source["coauthors"], ContributionParticipation);
	        this.keywords = source["keywords"];
	        this.references = source["references"];
	        this.board_number = source["board_number"];
	        this.code = source["code"];
	        this.url = source["url"];
	        this.note = source["note"];
	        this.material = source["material"];
	        this.folders = this.convertValues(source["folders"], Folder);
	        this.allowed = source["allowed"];
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
	
	
	
	
	export class Event {
	    id: string;
	    title: string;
	    description: string;
	    startDate?: DateInfo;
	    endDate?: DateInfo;
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
	        this.startDate = this.convertValues(source["startDate"], DateInfo);
	        this.endDate = this.convertValues(source["endDate"], DateInfo);
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

export namespace main {
	
	export class AppInfo {
	    name: string;
	    version: string;
	    author: string;
	    company: string;
	    authorEmail: string;
	    buildDate: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.author = source["author"];
	        this.company = source["company"];
	        this.authorEmail = source["authorEmail"];
	        this.buildDate = source["buildDate"];
	    }
	}

}

