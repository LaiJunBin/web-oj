export class Container {
	private g: Map<
		string,
		{ constructor: any; args: string[]; back: string[]; missing: Set<string> }
	> = new Map();
	private dirty = false;
	private instances: Map<string, any> = new Map();

	set<T>(constructor: new (...args: any) => T) {
		if (!constructor || typeof constructor !== 'function') {
			throw new Error('Invalid class, maybe it is not a class');
		}

		const key = constructor.name;
		if (!key) {
			throw new Error('Invalid class, maybe it is not a class');
		}

		// In production, __deps is injected by the diDepsPlugin build plugin.
		// In development, fall back to parsing constructor.toString().
		const args =
			(constructor as any).__deps ??
			/constructor\(([^)]*)\)/
				.exec(constructor.toString())?.[1]
				.split(',')
				.map((arg) => arg.trim())
				.filter((arg) => arg)
				.map((arg) => arg.charAt(0).toUpperCase() + arg.slice(1)) ??
			[];

		if ((args.length || 0) !== constructor.length) {
			throw new Error(
				`Invalid constructor arguments, expected ${constructor.length} but got ${args.length} in ${key}`
			);
		}

		const missingArgs = args.filter((arg: any) => !this.instances.has(arg));
		if (this.g.has(key)) {
			this.g.set(key, { ...this.g.get(key)!, constructor, args, missing: new Set(missingArgs) });
		} else {
			this.g.set(key, { constructor, args, back: [], missing: new Set(missingArgs) });
		}

		for (const arg of args) {
			if (this.g.has(arg)) {
				this.g.set(arg, { ...this.g.get(arg)!, back: [...this.g.get(arg)!.back, key] });
			} else {
				this.g.set(arg, { constructor: null, args: [], back: [key], missing: new Set() });
			}
		}

		this.dirty = true;
	}

	get<T extends new (...args: any) => any>(arg: T | string): InstanceType<T> {
		if (this.dirty) {
			this.updateInstances();
		}
		const key = typeof arg === 'string' ? arg : arg.name;
		const instance = this.instances.get(key);
		if (!instance) {
			throw new Error(`No instance found for ${key}`);
		}

		return instance;
	}

	private updateInstances() {
		this.dirty = false;

		const visited = new Set<string>();
		const queue: string[] = [];

		for (const key of this.g.keys()) {
			if (this.instances.has(key)) {
				continue;
			}
			const { missing } = this.g.get(key)!;
			if (missing.size === 0) {
				queue.push(key);
			}
		}

		while (queue.length) {
			const key = queue.shift()!;
			if (visited.has(key)) {
				continue;
			}
			visited.add(key);
			const { constructor, args, back } = this.g.get(key)!;
			if (!constructor) {
				throw new Error(`No constructor found for ${key}`);
			}

			const a = args.map((arg) => this.instances.get(arg));
			this.instances.set(key, new constructor(...a));
			for (const b of back) {
				const { missing } = this.g.get(b)!;
				missing.delete(key);
				if (missing.size === 0) {
					queue.push(b);
				}
			}
		}
	}
}
