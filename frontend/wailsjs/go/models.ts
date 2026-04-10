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
	    method_material_id: number;
	    name: string;
	    unit: string;
	    display_order: number;
	    render_chart: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ComboAnalyte(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mma_id = source["mma_id"];
	        this.method_material_id = source["method_material_id"];
	        this.name = source["name"];
	        this.unit = source["unit"];
	        this.display_order = source["display_order"];
	        this.render_chart = source["render_chart"];
	    }
	}
	export class CommentResponse {
	    id: number;
	    control_chart_id: number;
	    measurement_id?: number;
	    text: string;
	    user_id: number;
	    username: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new CommentResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.control_chart_id = source["control_chart_id"];
	        this.measurement_id = source["measurement_id"];
	        this.text = source["text"];
	        this.user_id = source["user_id"];
	        this.username = source["username"];
	        this.created_at = source["created_at"];
	    }
	}
	export class ControlLimitRegion {
	    id: number;
	    mma_id: number;
	    mean: number;
	    ucl: number;
	    lcl: number;
	    uwl?: number;
	    lwl?: number;
	    uil?: number;
	    lil?: number;
	    effective_from_sequence: number;
	    created_by: number;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ControlLimitRegion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.mma_id = source["mma_id"];
	        this.mean = source["mean"];
	        this.ucl = source["ucl"];
	        this.lcl = source["lcl"];
	        this.uwl = source["uwl"];
	        this.lwl = source["lwl"];
	        this.uil = source["uil"];
	        this.lil = source["lil"];
	        this.effective_from_sequence = source["effective_from_sequence"];
	        this.created_by = source["created_by"];
	        this.created_at = source["created_at"];
	    }
	}
	export class MMAEntry {
	    id: number;
	    method_material_id: number;
	    material_id: number;
	    material_name: string;
	    method_id: number;
	    method_name: string;
	    analyte_id: number;
	    analyte_name: string;
	    unit: string;
	    display_order: number;
	    render_chart: boolean;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MMAEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.method_material_id = source["method_material_id"];
	        this.material_id = source["material_id"];
	        this.material_name = source["material_name"];
	        this.method_id = source["method_id"];
	        this.method_name = source["method_name"];
	        this.analyte_id = source["analyte_id"];
	        this.analyte_name = source["analyte_name"];
	        this.unit = source["unit"];
	        this.display_order = source["display_order"];
	        this.render_chart = source["render_chart"];
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
	    method_material_id: number;
	
	    static createFrom(source: any = {}) {
	        return new MaterialSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.method_material_id = source["method_material_id"];
	    }
	}
	export class MeasurementResult {
	    mma_id: number;
	    analyte_name: string;
	    unit: string;
	    value: number;
	    sequence_number: number;
	    ucl?: number;
	    lcl?: number;
	    pass: boolean;
	    no_limits: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MeasurementResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mma_id = source["mma_id"];
	        this.analyte_name = source["analyte_name"];
	        this.unit = source["unit"];
	        this.value = source["value"];
	        this.sequence_number = source["sequence_number"];
	        this.ucl = source["ucl"];
	        this.lcl = source["lcl"];
	        this.pass = source["pass"];
	        this.no_limits = source["no_limits"];
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
	export class SPCRuleSet {
	    id: number;
	    methodMaterialId?: number;
	    effectiveFromSequence?: number;
	    beyondLimitsEnabled: boolean;
	    warningLimitsEnabled: boolean;
	    warningConsecutiveCount: number;
	    warningTriggerCount: number;
	    trendEnabled: boolean;
	    trendConsecutiveCount: number;
	    oneSideEnabled: boolean;
	    oneSideConsecutiveCount: number;
	    createdBy: number;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SPCRuleSet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.methodMaterialId = source["methodMaterialId"];
	        this.effectiveFromSequence = source["effectiveFromSequence"];
	        this.beyondLimitsEnabled = source["beyondLimitsEnabled"];
	        this.warningLimitsEnabled = source["warningLimitsEnabled"];
	        this.warningConsecutiveCount = source["warningConsecutiveCount"];
	        this.warningTriggerCount = source["warningTriggerCount"];
	        this.trendEnabled = source["trendEnabled"];
	        this.trendConsecutiveCount = source["trendConsecutiveCount"];
	        this.oneSideEnabled = source["oneSideEnabled"];
	        this.oneSideConsecutiveCount = source["oneSideConsecutiveCount"];
	        this.createdBy = source["createdBy"];
	        this.createdAt = source["createdAt"];
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

