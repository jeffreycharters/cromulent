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

