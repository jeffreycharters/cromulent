export namespace models {
	
	export class Analyte {
	    id: number;
	    name: string;
	    unit: string;
	
	    static createFrom(source: any = {}) {
	        return new Analyte(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.unit = source["unit"];
	    }
	}
	export class ComboAnalyte {
	    mma_id: number;
	    name: string;
	    unit: string;
	    display_order: number;
	
	    static createFrom(source: any = {}) {
	        return new ComboAnalyte(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mma_id = source["mma_id"];
	        this.name = source["name"];
	        this.unit = source["unit"];
	        this.display_order = source["display_order"];
	    }
	}
	export class MMAEntry {
	    id: number;
	    material_id: number;
	    material_name: string;
	    method_id: number;
	    method_name: string;
	    analyte_id: number;
	    analyte_name: string;
	    unit: string;
	    display_order: number;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MMAEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.material_id = source["material_id"];
	        this.material_name = source["material_name"];
	        this.method_id = source["method_id"];
	        this.method_name = source["method_name"];
	        this.analyte_id = source["analyte_id"];
	        this.analyte_name = source["analyte_name"];
	        this.unit = source["unit"];
	        this.display_order = source["display_order"];
	        this.active = source["active"];
	    }
	}
	export class Material {
	    id: number;
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new Material(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class MaterialSummary {
	    id: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new MaterialSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class Method {
	    id: number;
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new Method(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class MethodWithMaterials {
	    id: number;
	    name: string;
	    materials: MaterialSummary[];
	
	    static createFrom(source: any = {}) {
	        return new MethodWithMaterials(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.materials = this.convertValues(source["materials"], MaterialSummary);
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
	export class UserResponse {
	    id: number;
	    username: string;
	    role: string;
	    active: boolean;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new UserResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.role = source["role"];
	        this.active = source["active"];
	        this.created_at = source["created_at"];
	    }
	}

}

