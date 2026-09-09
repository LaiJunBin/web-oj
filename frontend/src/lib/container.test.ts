import {  describe, expect, test } from 'vitest';
import { Container } from './container';

describe('Container', () => {
	test('basic usage with simple set and get', () => {
		class A {}
		const container = new Container();
		container.set(A);
		const a = container.get(A);
		expect(a).toBeInstanceOf(A);
	});

	test('basic usage with extended class', () => {
		class A {}
		class B extends A {}
		const container = new Container();
		container.set(A);
		container.set(B);
		const b = container.get(B);
		expect(b).toBeInstanceOf(B);
	});

	test('extend class in args', () => {
		class A {}
		class B extends A {}
		class C {
			constructor(public b: B) {}
		}
		const container = new Container();
		container.set(A);
		container.set(B);
		container.set(C);
		const c = container.get(C);
		expect(c).toBeInstanceOf(C);
		expect(c.b).toBeInstanceOf(B);
	});

	test('singleton', () => {
		class A {}
		const container = new Container();
		container.set(A);
		const a = container.get(A);
		const b = container.get(A);
		expect(a).toBe(b);
	});

	test('two dependencies', () => {
		class A {}
		class B {}
		class C {
			constructor(a: A, b: B) {}
		}
		const container = new Container();
		container.set(A);
		container.set(B);
		container.set(C);
		const c = container.get(C);
		expect(c).toBeInstanceOf(C);
	});

	test('three depth dependencies and it should work', () => {
		class A {
			value = 1;
		}
		class B {
			constructor(public a: A) {}
		}
		class C {
			constructor(public b: B) {}
		}
		class D {
			constructor(public c: C) {}
		}
		const container = new Container();
		container.set(A);
		container.set(B);
		container.set(C);
		container.set(D);
		const d = container.get(D);
		expect(d).toBeInstanceOf(D);
		expect(d.c.b.a.value).toBe(1);
	});

	test('complex dependencies', () => {
		class A {}
		class B {
			constructor(public a: A) {}
		}
		class C {
			constructor(
				public a: A,
				public b: B
			) {}
		}
		class D {
			constructor(
				public a: A,
				public c: C
			) {}
		}
		class E {
			constructor(
				public a: A,
				public b: B,
				public c: C,
				public d: D
			) {}
		}

		const container = new Container();
		container.set(A);
		container.set(C);
		container.set(B);
		container.set(E);
		container.set(D);

		const instance = container.get(E);
		expect(instance).toBeInstanceOf(E);
		expect(instance.a).toBeInstanceOf(A);
		expect(instance.b).toBeInstanceOf(B);
		expect(instance.c).toBeInstanceOf(C);
		expect(instance.d).toBeInstanceOf(D);
	});

	test('lazy update container instance', () => {
		class A {}
		class B {
			constructor(public a: A) {}
		}
		const container = new Container();
		container.set(B);
		container.set(A);
		const instanceB = container.get(B);

		class C {
			constructor(public a: A) {}
		}
		container.set(C);
		const instanceC = container.get(C);
		expect(instanceB.a).toBe(instanceC.a);
	});

	test('should throw error if class is not a class', () => {
		const container = new Container();
		expect(() => container.set({} as any)).toThrowError('Invalid class, maybe it is not a class');
	});

	test('should throw error if constructor arguments are not matched', () => {
		class Arg {}
		class A {
			constructor(arg: Arg) {}
		}
		const container = new Container();
		container.set(A);
		expect(() => container.get(A)).toThrowError('No constructor found for Arg');
	});

	test('should throw error if constructor arguments are not matched', () => {
		class A {
			constructor(str: string) {}
		}
		const container = new Container();
		container.set(A);
		expect(() => container.get(A)).toThrowError('No constructor found for Str');
	});

	test('should throw error if instance not found', () => {
		class A {}
		const container = new Container();
		expect(() => container.get(A)).toThrowError('No instance found for A');
	});
});
