import { datetimeString, fileSize } from '@/lib';
import { describe, expect, test } from 'vitest';

describe('datetimeString', () => {
	const testcases = [
		{ date: null, defaultText: 'default text', expected: 'default text' },
		{ date: 'invalid date', defaultText: 'default text', expected: 'default text' },
		{ date: new Date('2023-10-05T14:48:00'), expected: '2023-10-05 14:48:00' },
		{ date: '2023-10-05T14:48:00', expected: '2023-10-05 14:48:00' },
		{ date: 1696512480000, expected: '2023-10-05 21:28:00' },
		{ date: undefined, defaultText: 'default text', expected: 'default text' },
		{ date: new Date('2023-10-05T14:48:07'), expected: '2023-10-05 14:48:07' },
		{ date: new Date('2023-10-05T14:07:48'), expected: '2023-10-05 14:07:48' },
		{ date: new Date('2023-10-05T07:48:48'), expected: '2023-10-05 07:48:48' },
		{ date: new Date('2023-10-05T14:48:48'), expected: '2023-10-05 14:48:48' },
		{ date: new Date('2023-10-15T14:48:48'), expected: '2023-10-15 14:48:48' }
	];

	testcases.forEach(({ date, defaultText, expected }) => {
		test(`when date is ${date}, it should return ${expected}`, () => {
			const result = datetimeString(date, defaultText);
			expect(result).toBe(expected);
		});
	});
});

describe('fileSize', () => {
	const testcases = [
		{ bytes: 0, expected: '0 Byte' },
		{ bytes: 512, expected: '512 Bytes' },
		{ bytes: 1024, expected: '1 KB' },
		{ bytes: 1048576, expected: '1 MB' },
		{ bytes: 1073741824, expected: '1 GB' },
		{ bytes: 1099511627776, expected: '1 TB' },
		{ bytes: 123456789, expected: '117.74 MB' },
		{ bytes: 1536, expected: '1.5 KB' },
		{ bytes: 1572864, expected: '1.5 MB' },
		{ bytes: 1610612736, expected: '1.5 GB' },
		{ bytes: 1649267441664, expected: '1.5 TB' }
	];

	testcases.forEach(({ bytes, expected }) => {
		test(`when bytes is ${bytes}, it should return ${expected}`, () => {
			const result = fileSize(bytes);
			expect(result).toBe(expected);
		});
	});
});
