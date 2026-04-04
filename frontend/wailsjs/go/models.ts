export namespace models {
	
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

